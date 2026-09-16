package fs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"unicode/utf8"

	"github.com/photoprism/photoprism/pkg/rnd"
)

// linkFile publishes a staged file under a name it must not replace. A test replaces it to take the
// path of a file system that has no hard links.
var linkFile = os.Link

// destExists reports the error a destination that may not be replaced returns.
func destExists(dest string) error {
	return fmt.Errorf("destination %s already exists: %w", filepath.Base(dest), os.ErrExist)
}

// destReplaceable reports whether a destination may be replaced without the force flag, which an
// empty regular file may be. It reads the name itself rather than what it resolves to.
func destReplaceable(dest string) bool {
	info, err := os.Lstat(dest)

	return err == nil && info.Mode().IsRegular() && info.Size() == 0
}

// checkDest reports whether a destination may be written, given the force flag.
func checkDest(dest string, force bool) error {
	// Links to files and directories inside the library are a supported layout, so a call leaves one
	// to its owner: it writes neither through a link nor over it, and the target and whatever keeps
	// it are left as they are.
	if IsSymlink(dest) {
		return fmt.Errorf("destination %s is a symbolic link", filepath.Base(dest))
	}

	if force || destReplaceable(dest) {
		return nil
	}

	if _, err := os.Lstat(dest); err == nil {
		return destExists(dest)
	}

	return nil
}

// stageName returns a hidden, uniquely named sibling of the destination that an indexing pass
// ignores. It keeps the destination's extension, which the media tools that read a staged file
// detect its type from, and clamps the base so the name still fits a directory entry.
func stageName(dest string) string {
	const maxNameLen = 255

	dir, base := filepath.Split(dest)
	ext := filepath.Ext(base)

	// A name whose extension alone would fill the entry is staged without it.
	if len(ext) > maxNameLen/2 {
		ext = ""
	}

	prefix, suffix := ".", "."+rnd.Base36(8)+ExtTmp+ext

	// Clamped on a rune boundary, since a file system that validates names refuses a split character.
	if limit := maxNameLen - len(prefix) - len(suffix); len(base) > limit {
		base = base[:limit]

		for len(base) > 0 && !utf8.ValidString(base) {
			base = base[:len(base)-1]
		}
	}

	return dir + prefix + base + suffix
}

// OpenStageFile creates an exclusive temporary sibling with the default file creation mode.
// The caller owns the handle and must remove or publish its pathname.
func OpenStageFile(dest string) (*os.File, error) {
	return os.OpenFile(stageName(dest), os.O_WRONLY|os.O_CREATE|os.O_EXCL, ModeFile) //nolint:gosec // the name is derived from a validated destination
}

// CreateStageFile creates an empty, uniquely named sibling of the destination and returns its name.
// A caller writing through a subprocess needs the name reserved before it starts, so the create is
// what makes it theirs; it publishes with PublishFile and removes the file on every other way out.
// The directory must already exist, and the staged file gets ModeFile rather than the destination's.
func CreateStageFile(dest string) (name string, err error) {
	f, err := OpenStageFile(dest)

	if err != nil {
		return "", err
	}

	name = f.Name()

	if err = f.Close(); err != nil {
		_ = os.Remove(name)

		return "", err
	}

	return name, nil
}

// PublishFile publishes a staged file with the shared Copy/Move replacement rules. An empty regular
// destination may be replaced without force; filesystems without hard links use a checked rename.
// Preflight checks and metadata preservation belong to callers, so a forced publish replaces a
// destination symlink rather than refusing it.
func PublishFile(staged, dest string, force bool) error {
	return publishFile(staged, dest, force)
}

// publishFile puts a staged file in place of the destination. A call that may replace the destination
// renames onto it, and one that may not links, since a link fails when the name is taken; a file
// system without links is checked with a stat instead.
func publishFile(staged, dest string, force bool) error {
	if force {
		return os.Rename(staged, dest)
	}

	err := linkFile(staged, dest)

	switch {
	case err == nil:
		// The staged file is published under its new name, so removing the old one cannot fail the
		// operation.
		_ = os.Remove(staged)

		return nil
	case errors.Is(err, os.ErrExist):
		if destReplaceable(dest) {
			return os.Rename(staged, dest)
		}

		return destExists(dest)
	}

	if _, statErr := os.Lstat(dest); statErr == nil {
		if destReplaceable(dest) {
			return os.Rename(staged, dest)
		}

		return destExists(dest)
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return errors.Join(err, statErr)
	}

	return os.Rename(staged, dest)
}

// Copy copies a file to a destination.
func Copy(src, dest string, force bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s (panic)", r)
		}
	}()

	// Check for obviously empty or invalid source and destination file paths.
	if src == "" || src == "." || src == ".." {
		return errors.New("invalid copy source file path")
	} else if dest == "" || dest == "." || dest == ".." {
		return errors.New("invalid copy destination file path")
	}

	// Check whether a destination file and directory name are specified.
	if filepath.Base(dest) == "" {
		return errors.New("invalid copy destination name")
	} else if filepath.Dir(dest) == "" {
		return errors.New("invalid copy destination path")
	}

	// Resolve absolute destination file path and return an error if unsuccessful.
	if dest, err = filepath.Abs(dest); err != nil {
		return err
	}

	destDir := filepath.Dir(dest)

	// Error if source and destination file path are the same.
	if dest == src {
		return fmt.Errorf("cannot copy file %s onto itself", filepath.Base(dest))
	}

	// Report a destination that may not be written before the source is read.
	if err = checkDest(dest, force); err != nil {
		return err
	}

	// Make sure the target directory exists.
	if err = MkdirAll(destDir); err != nil {
		return err
	}

	srcFile, err := os.Open(src) //nolint:gosec // src is validated by callers

	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, srcFile.Close())
	}()

	// Write through a staged sibling, so the destination is never opened by name.
	destFile, err := OpenStageFile(dest)

	if err != nil {
		return err
	}

	// Publishing replaces the destination, so what a replaced one carried on its own inode is
	// carried over first. Ownership needs the privilege to set it and is restored only where the
	// process has it; extended attributes and access control lists are not carried at all.
	if info, statErr := os.Stat(dest); statErr == nil {
		_ = destFile.Chmod(info.Mode().Perm())

		if owner, ok := info.Sys().(*syscall.Stat_t); ok {
			_ = destFile.Chown(int(owner.Uid), int(owner.Gid))
		}
	}

	staged := destFile.Name()
	published := false

	// Remove only the file this call created, on every way out including a panic.
	defer func() {
		if !published {
			err = errors.Join(err, os.Remove(staged))
		}
	}()

	buf := getCopyBuffer()
	defer putCopyBuffer(buf)

	_, err = io.CopyBuffer(destFile, srcFile, buf)

	if closeErr := destFile.Close(); err == nil {
		err = closeErr
	}

	if err != nil {
		return err
	}

	if err = publishFile(staged, dest, force); err != nil {
		return err
	}

	published = true

	return nil
}

// Move moves an existing file to a new destination and returns an error if it fails.
func Move(src, dest string, force bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s (panic)", r)
		}
	}()

	// Check for obviously empty or invalid source and destination file paths.
	if src == "" || src == "." || src == ".." {
		return errors.New("invalid move source file path")
	} else if dest == "" || dest == "." || dest == ".." {
		return errors.New("invalid move destination file path")
	}

	// Check whether a destination file and directory name are specified.
	if filepath.Base(dest) == "" {
		return errors.New("invalid move destination name")
	} else if filepath.Dir(dest) == "" {
		return errors.New("invalid move destination path")
	}

	// Resolve absolute destination file path and return an error if unsuccessful.
	if dest, err = filepath.Abs(dest); err != nil {
		return err
	}

	destDir := filepath.Dir(dest)

	// Error if source and destination file path are the same.
	if dest == src {
		return fmt.Errorf("cannot move file %s onto itself", filepath.Base(dest))
	}

	// Report a destination that may not be written before anything is moved.
	if err = checkDest(dest, force); err != nil {
		return err
	}

	// Make sure the target directory exists.
	if err = MkdirAll(destDir); err != nil {
		return err
	}

	// A rename replaces the name rather than what it resolves to, so it publishes the move on its
	// own; without the force flag a link does the same and refuses a name that is taken.
	if force {
		if err = os.Rename(src, dest); err == nil {
			return nil
		}
	} else if err = linkFile(src, dest); err == nil {
		return os.Remove(src)
	} else if errors.Is(err, os.ErrExist) {
		if !destReplaceable(dest) {
			return destExists(dest)
		}

		if err = os.Rename(src, dest); err == nil {
			return nil
		}
	} else if _, statErr := os.Lstat(dest); errors.Is(statErr, os.ErrNotExist) {
		// A file system without hard links keeps the rename it would otherwise have taken.
		if err = os.Rename(src, dest); err == nil {
			return nil
		}
	}

	// Separate mounts for originals and import are an ordinary layout, so a rename or link across
	// devices falls back to a copy.
	if err = Copy(src, dest, force); err != nil {
		return err
	}

	return os.Remove(src)
}
