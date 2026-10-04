package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
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
	return t.findEach(fileName, dirs, baseDir, stripSequence, false, accept)
}

// FindGenerated searches like FindEach, except that in directories other than the folder of the file it only
// checks the names PhotoPrism generates: the full name and BasePrefix with the default extension of the type.
// In the folder of the file, it also checks the full name in other cases, see fullNameCases.
func (t Type) FindGenerated(fileName string, dirs []string, baseDir string, stripSequence bool, accept func(fileName string) bool) string {
	return t.findEach(fileName, dirs, baseDir, stripSequence, true, accept)
}

// findEach implements FindEach and FindGenerated. Each directory and name variant is checked once.
func (t Type) findEach(fileName string, dirs []string, baseDir string, stripSequence, generated bool, accept func(fileName string) bool) string {
	fileBasePrefix := BasePrefix(fileName, stripSequence)

	// A name without a base would match files named after the folder.
	if NoBaseName(fileBasePrefix) {
		return ""
	}

	generatedNames := appendUnique(nil, filepath.Base(fileName), fileBasePrefix)
	names := generatedNames

	if !ignoreCase {
		names = appendUnique(slices.Clone(names), strings.ToLower(fileBasePrefix), strings.ToUpper(fileBasePrefix))
	}

	ownNames := names

	if generated && !ignoreCase {
		ownNames = appendUnique(slices.Clone(names), fullNameCases(fileName)...)
	}

	filePath := filepath.Dir(fileName)
	search := []string{filePath}
	lastDir := filePath

	for _, dir := range dirs {
		if dir == "" || dir == lastDir {
			continue
		}

		lastDir = dir

		switch {
		case dir == filePath:
			continue
		case filepath.IsAbs(dir):
			dir = filepath.Join(dir, RelName(filePath, baseDir))
		default:
			dir = filepath.Join(filePath, dir)
		}

		search = appendUnique(search, dir)
	}

	defaultExt := t.DefaultExt()

	for _, ext := range FileTypes[t] {
		for i, dir := range search {
			candidates := names

			switch {
			case i == 0:
				candidates = ownNames
			case generated && ext != defaultExt:
				continue
			case generated:
				candidates = generatedNames
			}

			for _, name := range candidates {
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

// fullNameCases returns the full base name with its name and its extensions each as given, in lower case, or in
// upper case, e.g. "img_1234.RAW" for "IMG_1234.raw".
func fullNameCases(fileName string) (result []string) {
	fullName := filepath.Base(fileName)
	name := BasePrefix(fileName, false)
	exts := strings.TrimPrefix(fullName, name)

	for _, n := range []string{name, strings.ToLower(name), strings.ToUpper(name)} {
		result = appendUnique(result, n+exts, n+strings.ToLower(exts), n+strings.ToUpper(exts))
	}

	return result
}

// appendUnique appends the values that the list does not contain yet.
func appendUnique(list []string, values ...string) []string {
	for _, v := range values {
		if !slices.Contains(list, v) {
			list = append(list, v)
		}
	}

	return list
}
