package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Format represents a file format type.
type Format string

// String returns the file format as string.
func (f Format) String() string {
	return string(f)
}

// Is checks if the format strings match.
func (f Format) Is(s string) bool {
	if s == "" {
		return false
	}

	return f.String() == strings.ToLower(s)
}

// Ext returns the standard file format extension.
func (f Format) Ext() string {
	return fmt.Sprintf(".%s", f)
}

// Find returns the first filename with the same base name and a given type.
func (f Format) Find(fileName string, stripSequence bool) string {
	base := BasePrefix(fileName, stripSequence)
	dir := filepath.Dir(fileName)

	prefix := filepath.Join(dir, base)
	prefixLower := filepath.Join(dir, strings.ToLower(base))
	prefixUpper := filepath.Join(dir, strings.ToUpper(base))

	for _, ext := range Formats[f] {
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
func (f Format) FindFirst(fileName string, dirs []string, baseDir string, stripSequence bool) string {
	fileBase := filepath.Base(fileName)
	fileBasePrefix := BasePrefix(fileName, stripSequence)
	fileBaseLower := strings.ToLower(fileBasePrefix)
	fileBaseUpper := strings.ToUpper(fileBasePrefix)

	fileDir := filepath.Dir(fileName)
	search := append([]string{fileDir}, dirs...)

	for _, ext := range Formats[f] {
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
			} else if info, err := os.Stat(filepath.Join(dir, fileBasePrefix) + ext); err == nil && info.Mode().IsRegular() {
				return filepath.Join(dir, info.Name())
			}

			if ignoreCase {
				continue
			}

			if info, err := os.Stat(filepath.Join(dir, fileBaseLower) + ext); err == nil && info.Mode().IsRegular() {
				return filepath.Join(dir, info.Name())
			} else if info, err := os.Stat(filepath.Join(dir, fileBaseUpper) + ext); err == nil && info.Mode().IsRegular() {
				return filepath.Join(dir, info.Name())
			}
		}
	}

	return ""
}

// FindAll searches a list of directories for files with the same base name and a given type.
func (f Format) FindAll(fileName string, dirs []string, baseDir string, stripSequence bool) (results []string) {
	fileBase := filepath.Base(fileName)
	fileBasePrefix := BasePrefix(fileName, stripSequence)
	fileBaseLower := strings.ToLower(fileBasePrefix)
	fileBaseUpper := strings.ToUpper(fileBasePrefix)

	fileDir := filepath.Dir(fileName)
	search := append([]string{fileDir}, dirs...)

	for _, ext := range Formats[f] {
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

// FileFormat returns the (expected) type for a given file name.
func FileFormat(fileName string) Format {
	fileExt := strings.ToLower(filepath.Ext(fileName))
	result, ok := Extensions[fileExt]

	if !ok {
		result = FormatOther
	}

	return result
}
