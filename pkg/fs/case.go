package fs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
)

const (
	// caseProbeDirs is the number of folders CaseInsensitiveDir reads at most.
	caseProbeDirs = 4
	// caseProbeEntries is the number of entries CaseInsensitiveDir reads per folder.
	caseProbeEntries = 64
	// caseProbeDepth is the number of folder levels CaseInsensitiveDir reads, including the root.
	caseProbeDepth = 3
)

// caseScopes contains the folders whose lookups follow their own case mode rather than ignoreCase, the most
// specific first, e.g. the originals path and the file systems mounted below it.
var caseScopes []CaseScope

// fileTypesAll contains each extension followed by its uppercase variant, the search order of FileTypes
// in case-sensitive mode.
var fileTypesAll = ExtensionList.Types(false)

// CaseInsensitive tests if a storage path is case-insensitive.
func CaseInsensitive(storagePath string) (result bool, err error) {
	tmpName := filepath.Join(storagePath, ".caseTest.tmp")

	if err = os.WriteFile(tmpName, []byte("{}"), ModeFile); err != nil {
		return false, fmt.Errorf("%s not writable", filepath.Base(storagePath))
	}

	defer func() {
		_ = os.Remove(tmpName)
	}()

	result = FileExists(filepath.Join(storagePath, ".CASETEST.TMP"))

	return result, err
}

// IgnoreCase enables the case-insensitive mode for lookups outside the folders set with SetCaseScopes.
func IgnoreCase() {
	SetIgnoreCase(true)
}

// SetIgnoreCase sets the case mode for lookups outside the folders set with SetCaseScopes.
func SetIgnoreCase(enabled bool) {
	ignoreCase = enabled
	FileTypes = ExtensionList.Types(enabled)
}

// CaseScope describes whether lookups in Dir and below are case-insensitive.
type CaseScope struct {
	Dir    string
	Ignore bool
}

// SetCaseScopes replaces the folders whose lookups follow their own case mode, regardless of IgnoreCase. In
// nested folders, the most specific one applies; for the same folder, the last one given.
func SetCaseScopes(scopes ...CaseScope) {
	result := make([]CaseScope, 0, len(scopes))

	for _, s := range scopes {
		if s.Dir == "" {
			continue
		}

		s.Dir = filepath.Clean(s.Dir)
		result = slices.DeleteFunc(result, func(r CaseScope) bool { return r.Dir == s.Dir })
		result = append(result, s)
	}

	// Sort the most specific first; for the same length, case-sensitive first.
	slices.SortStableFunc(result, func(a, b CaseScope) int {
		if n := len(b.Dir) - len(a.Dir); n != 0 {
			return n
		} else if a.Ignore == b.Ignore {
			return 0
		} else if a.Ignore {
			return 1
		}

		return -1
	})

	caseScopes = result
}

// CaseMode represents the case-insensitive lookup settings, so that tests can restore them.
type CaseMode struct {
	ignoreCase bool
	caseScopes []CaseScope
	fileTypes  TypesExt
}

// GetCaseMode returns the current case-insensitive lookup settings.
func GetCaseMode() CaseMode {
	return CaseMode{ignoreCase: ignoreCase, caseScopes: caseScopes, fileTypes: FileTypes}
}

// RestoreCaseMode restores settings returned by GetCaseMode.
func RestoreCaseMode(m CaseMode) {
	ignoreCase, caseScopes, FileTypes = m.ignoreCase, m.caseScopes, m.fileTypes
}

// ignoreCaseIn reports whether lookups in dir are case-insensitive. Case-sensitive scopes match regardless of
// letter case, as a case-insensitive parent file system may list a mount point under another spelling.
func ignoreCaseIn(dir string) bool {
	for _, s := range caseScopes {
		if s.Ignore && InDir(dir, s.Dir) || !s.Ignore && inDirFold(dir, s.Dir) {
			return s.Ignore
		}
	}

	return ignoreCase
}

// inDirFold works like InDir, but compares letters case-insensitively.
func inDirFold(fileName, dir string) bool {
	switch {
	case fileName == "" || dir == "" || len(fileName) < len(dir):
		return false
	case len(fileName) == len(dir):
		return strings.EqualFold(fileName, dir)
	case strings.HasSuffix(dir, string(os.PathSeparator)):
		return strings.EqualFold(fileName[:len(dir)], dir)
	default:
		return fileName[len(dir)] == os.PathSeparator && strings.EqualFold(fileName[:len(dir)], dir)
	}
}

// CaseInsensitiveDir tests if dir is on a case-insensitive file system without writing to it, and returns an
// error naming the reason if it cannot tell. It looks up a file name with its ASCII letter case swapped, reading
// at most caseProbeDirs folders on the same device without following links below dir, and does not open the
// paths in skip, e.g. mount points. A CIFS mount with noserverino reads as unknown.
func CaseInsensitiveDir(dir string, skip ...string) (insensitive bool, err error) {
	return caseProbe{readDir: readDirN, lstat: os.Lstat, skip: skip}.run(dir)
}

// errOtherDevice is returned by readDirN for a folder on another device than the root.
var errOtherDevice = errors.New("other device")

// Errors returned by CaseInsensitiveDir name the reason why the result is unknown, without the path.
var (
	errCaseRead      = errors.New("folder not readable")
	errCaseLookup    = errors.New("file name lookup failed")
	errCaseOtherFile = errors.New("file name with swapped case is another file")
	errCaseNoName    = errors.New("no file name with ASCII letters within reach")
)

// anyDevice makes readDirN read a folder on any device.
const anyDevice = ^uint64(0)

// caseProbe implements CaseInsensitiveDir with replaceable file system calls.
type caseProbe struct {
	readDir func(dir string, n int, dev uint64) ([]os.DirEntry, uint64, error)
	lstat   func(name string) (os.FileInfo, error)
	skip    []string
}

// run checks the first regular file in dir, or in a folder below it, whose name changes with its letter case
// swapped. Both spellings in the listing, or a missing swapped name, mean case-sensitive, and the same file means
// case-insensitive; a different file, an error, or no such file within the budget leave it unknown.
func (p caseProbe) run(dir string) (insensitive bool, err error) {
	type folder struct {
		path  string
		depth int
	}

	if dir == "" {
		return false, errCaseRead
	}

	rootDev := anyDevice
	stack := []folder{{path: filepath.Clean(dir), depth: 1}}

	for reads := 0; len(stack) > 0 && reads < caseProbeDirs; reads++ {
		f := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		entries, dev, readErr := p.readDir(f.path, caseProbeEntries, rootDev)

		switch {
		case readErr != nil && reads == 0:
			return false, errCaseRead
		case readErr != nil:
			continue
		case reads == 0:
			rootDev = dev
		}

		names := make([]string, len(entries))

		for i, e := range entries {
			names[i] = e.Name()
		}

		for i, name := range names {
			swapped := swapASCIICase(name)

			// Only regular files are looked up, as a CIFS client stops using server inode numbers for
			// the whole mount when it finds a folder under a second spelling.
			if swapped == name || !entries[i].Type().IsRegular() || p.skipped(filepath.Join(f.path, name)) {
				continue
			} else if slices.Contains(names, swapped) {
				return false, nil
			}

			info, lstatErr := p.lstat(filepath.Join(f.path, name))

			if lstatErr != nil {
				return false, errCaseLookup
			}

			swappedInfo, lstatErr := p.lstat(filepath.Join(f.path, swapped))

			switch {
			case errors.Is(lstatErr, os.ErrNotExist):
				return false, nil
			case lstatErr != nil:
				return false, errCaseLookup
			case os.SameFile(info, swappedInfo):
				return true, nil
			default:
				return false, errCaseOtherFile
			}
		}

		if f.depth >= caseProbeDepth {
			continue
		}

		// Push in reverse, so that the first subfolder listed is read next.
		for i := len(entries) - 1; i >= 0; i-- {
			if !entries[i].IsDir() {
				continue
			} else if sub := filepath.Join(f.path, entries[i].Name()); !p.skipped(sub) {
				stack = append(stack, folder{path: sub, depth: f.depth + 1})
			}
		}
	}

	return false, errCaseNoName
}

// skipped reports whether name is in skip, regardless of letter case.
func (p caseProbe) skipped(name string) bool {
	return slices.ContainsFunc(p.skip, func(s string) bool { return strings.EqualFold(s, name) })
}

// readDirN returns up to n entries of dir in directory order and the device number of dir. Unless dev is
// anyDevice, it opens only a directory that is not a link and returns errOtherDevice before reading if dir is
// on another device.
func readDirN(dir string, n int, dev uint64) ([]os.DirEntry, uint64, error) {
	flags := os.O_RDONLY | syscall.O_DIRECTORY | syscall.O_NONBLOCK

	if dev != anyDevice {
		flags |= syscall.O_NOFOLLOW
	}

	f, err := os.OpenFile(dir, flags, 0) //nolint:gosec // dir is a configured storage path or a folder below it

	if err != nil {
		return nil, 0, err
	}

	defer f.Close()

	info, err := f.Stat()

	if err != nil {
		return nil, 0, err
	} else if !info.IsDir() {
		return nil, 0, fmt.Errorf("%s is not a directory", filepath.Base(dir))
	}

	var dirDev uint64

	if st, isStat := info.Sys().(*syscall.Stat_t); isStat {
		dirDev = uint64(st.Dev) //nolint:unconvert // Dev is uint32 on some platforms
	}

	if dev != anyDevice && dirDev != dev {
		return nil, dirDev, errOtherDevice
	}

	entries, err := f.ReadDir(n)

	if errors.Is(err, io.EOF) {
		err = nil
	}

	return entries, dirDev, err
}

// swapASCIICase returns s with the case of its ASCII letters swapped.
func swapASCIICase(s string) string {
	b := []byte(s)

	for i, c := range b {
		switch {
		case c >= 'a' && c <= 'z':
			b[i] = c - ('a' - 'A')
		case c >= 'A' && c <= 'Z':
			b[i] = c + ('a' - 'A')
		}
	}

	return string(b)
}
