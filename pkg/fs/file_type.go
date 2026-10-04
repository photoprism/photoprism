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

// IsPreviewImageExt checks if the filename has a JPEG or PNG extension, the formats used to show images.
func IsPreviewImageExt(fileName string) bool {
	t := FileType(fileName)
	return t == ImageJpeg || t == ImagePng
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

// ToUpper returns the file format as uppercase string.
func (t Type) ToUpper() string {
	return strings.ToUpper(t.String())
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

	// A name without a base would match files named after the folder.
	if NoBaseName(base) {
		return ""
	}

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
	return t.FindEach(fileName, dirs, baseDir, stripSequence, nil)
}

// FindAll searches a list of directories for files with the same base name and a given type.
func (t Type) FindAll(fileName string, dirs []string, baseDir string, stripSequence bool) (results []string) {
	t.FindEach(fileName, dirs, baseDir, stripSequence, func(name string) bool {
		results = append(results, name)
		return false
	})

	return results
}

// FindEach searches a list of directories for files with the same base name and a given type, by extension,
// directory and name variant, and returns the first file that accept takes, or the first file found if accept
// is nil. It returns "" if there is none.
func (t Type) FindEach(fileName string, dirs []string, baseDir string, stripSequence bool, accept func(fileName string) bool) string {
	fileBasePrefix := BasePrefix(fileName, stripSequence)

	// A name without a base would match files named after the folder.
	if NoBaseName(fileBasePrefix) {
		return ""
	}

	names := []string{filepath.Base(fileName), fileBasePrefix}

	if !ignoreCase {
		names = append(names, strings.ToLower(fileBasePrefix), strings.ToUpper(fileBasePrefix))
	}

	filePath := filepath.Dir(fileName)
	search := make([]string, 0, len(dirs)+1)
	lastDir := ""

	for _, dir := range append([]string{filePath}, dirs...) {
		if dir == "" || dir == lastDir {
			continue
		}

		lastDir = dir

		if dir != filePath {
			if filepath.IsAbs(dir) {
				dir = filepath.Join(dir, RelName(filePath, baseDir))
			} else {
				dir = filepath.Join(filePath, dir)
			}
		}

		search = append(search, dir)
	}

	for _, ext := range FileTypes[t] {
		for _, dir := range search {
			for _, name := range names {
				if info, err := os.Stat(filepath.Join(dir, name) + ext); err != nil || !info.Mode().IsRegular() {
					continue
				} else if found := filepath.Join(dir, info.Name()); accept == nil || accept(found) {
					return found
				}
			}
		}
	}

	return ""
}
