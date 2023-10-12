package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileType returns the type associated with the specified filename,
// and TypeUnknown if it could not be matched.
func FileType(fileName string) Type {
	if t, found := Extensions[LowerExt(fileName)]; found {
		return t
	}

	return TypeUnknown
}

// IsAnimatedImage checks if the type associated with the specified filename may be animated.
func IsAnimatedImage(fileName string) bool {
	if t, found := Extensions[LowerExt(fileName)]; found {
		return TypeAnimated[t] != ""
	}

	return false
}

// NewType creates a new file type from a filename extension.
func NewType(ext string) Type {
	return Type(TrimExt(ext))
}

// Type represents a file format type.
type Type string

// String returns the file format as string.
func (t Type) String() string {
	return string(t)
}

// Equal checks if the type matches.
func (t Type) Equal(s string) bool {
	return strings.EqualFold(s, t.String())
}

// NotEqual checks if the type is different.
func (t Type) NotEqual(s string) bool {
	return !t.Equal(s)
}

// DefaultExt returns the default file format extension with dot.
func (t Type) DefaultExt() string {
	return fmt.Sprintf(".%s", t)
}

// Find returns the first filename with the same base name and a given type.
func (t Type) Find(fileName string, stripSequence bool) string {
	base := BasePrefix(fileName, stripSequence)
	dir := filepath.Dir(fileName)

	prefix := filepath.Join(dir, base)
	prefixLower := filepath.Join(dir, strings.ToLower(base))
	prefixUpper := filepath.Join(dir, strings.ToUpper(base))

	for _, ext := range FileTypes[t] {
		if info, err := os.Stat(prefix + ext); err == nil && info.Mode().IsRegular() {
			return filepath.Join(dir, info.Name())
		}

		if ignoreCase {
			continue
		}

		if info, err := os.Stat(prefixLower + ext); err == nil && info.Mode().IsRegular() {
			return filepath.Join(dir, info.Name())
		}

		if info, err := os.Stat(prefixUpper + ext); err == nil && info.Mode().IsRegular() {
			return filepath.Join(dir, info.Name())
		}
	}

	return ""
}

// FindFirst searches a list of directories for the first file with the same base name and a given type.
func (t Type) FindFirst(fileName string, dirs []string, baseDir string, stripSequence bool) string {
	fileBase := filepath.Base(fileName)
	fileBasePrefix := BasePrefix(fileName, stripSequence)
	fileBaseLower := strings.ToLower(fileBasePrefix)
	fileBaseUpper := strings.ToUpper(fileBasePrefix)

	fileDir := filepath.Dir(fileName)
	search := append([]string{fileDir}, dirs...)

	for _, ext := range FileTypes[t] {
		lastDir := ""

		for _, dir := range search {
			if dir == "" || dir == lastDir {
				continue
			}

			lastDir = dir

			if dir != fileDir {
				if filepath.IsAbs(dir) {
					dir = filepath.Join(dir, RelName(fileDir, baseDir))
				} else {
					dir = filepath.Join(fileDir, dir)
				}
			}

			if info, err := os.Stat(filepath.Join(dir, fileBase) + ext); err == nil && info.Mode().IsRegular() {
				return filepath.Join(dir, info.Name())
			} else if info, err = os.Stat(filepath.Join(dir, fileBasePrefix) + ext); err == nil && info.Mode().IsRegular() {
				return filepath.Join(dir, info.Name())
			}

			if ignoreCase {
				continue
			}

			if info, err := os.Stat(filepath.Join(dir, fileBaseLower) + ext); err == nil && info.Mode().IsRegular() {
				return filepath.Join(dir, info.Name())
			} else if info, err = os.Stat(filepath.Join(dir, fileBaseUpper) + ext); err == nil && info.Mode().IsRegular() {
				return filepath.Join(dir, info.Name())
			}
		}
	}

	return ""
}

// FindAll searches a list of directories for files with the same base name and a given type.
func (t Type) FindAll(fileName string, dirs []string, baseDir string, stripSequence bool) (results []string) {
	fileBase := filepath.Base(fileName)
	fileBasePrefix := BasePrefix(fileName, stripSequence)
	fileBaseLower := strings.ToLower(fileBasePrefix)
	fileBaseUpper := strings.ToUpper(fileBasePrefix)

	fileDir := filepath.Dir(fileName)
	search := append([]string{fileDir}, dirs...)

	for _, ext := range FileTypes[t] {
		lastDir := ""

		for _, dir := range search {
			if dir == "" || dir == lastDir {
				continue
			}

			lastDir = dir

			if dir != fileDir {
				if filepath.IsAbs(dir) {
					dir = filepath.Join(dir, RelName(fileDir, baseDir))
				} else {
					dir = filepath.Join(fileDir, dir)
				}
			}

			if info, err := os.Stat(filepath.Join(dir, fileBase) + ext); err == nil && info.Mode().IsRegular() {
				results = append(results, filepath.Join(dir, info.Name()))
			}

			if info, err := os.Stat(filepath.Join(dir, fileBasePrefix) + ext); err == nil && info.Mode().IsRegular() {
				results = append(results, filepath.Join(dir, info.Name()))
			}

			if ignoreCase {
				continue
			}

			if info, err := os.Stat(filepath.Join(dir, fileBaseLower) + ext); err == nil && info.Mode().IsRegular() {
				results = append(results, filepath.Join(dir, info.Name()))
			}

			if info, err := os.Stat(filepath.Join(dir, fileBaseUpper) + ext); err == nil && info.Mode().IsRegular() {
				results = append(results, filepath.Join(dir, info.Name()))
			}
		}
	}

	return results
}
