package fs

import (
	"os"
	"path/filepath"
	"strings"
)

// FileFormat represents a file format type.
type FileFormat string

// String returns the file format as string.
func (f FileFormat) String() string {
	return string(f)
}

// Is checks if the format strings match.
func (f FileFormat) Is(s string) bool {
	if s == "" {
		return false
	}

	return f.String() == strings.ToLower(s)
}

// Find returns the first filename with the same base name and a given type.
func (f FileFormat) Find(fileName string, stripSequence bool) string {
	base := BasePrefix(fileName, stripSequence)
	dir := filepath.Dir(fileName)

	prefix := filepath.Join(dir, base)
	prefixLower := filepath.Join(dir, strings.ToLower(base))
	prefixUpper := filepath.Join(dir, strings.ToUpper(base))

	for _, ext := range TypeExt[f] {
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
func (f FileFormat) FindFirst(fileName string, dirs []string, baseDir string, stripSequence bool) string {
	fileBase := filepath.Base(fileName)
	fileBasePrefix := BasePrefix(fileName, stripSequence)
	fileBaseLower := strings.ToLower(fileBasePrefix)
	fileBaseUpper := strings.ToUpper(fileBasePrefix)

	fileDir := filepath.Dir(fileName)
	search := append([]string{fileDir}, dirs...)

	for _, ext := range TypeExt[f] {
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
func (f FileFormat) FindAll(fileName string, dirs []string, baseDir string, stripSequence bool) (results []string) {
	fileBase := filepath.Base(fileName)
	fileBasePrefix := BasePrefix(fileName, stripSequence)
	fileBaseLower := strings.ToLower(fileBasePrefix)
	fileBaseUpper := strings.ToUpper(fileBasePrefix)

	fileDir := filepath.Dir(fileName)
	search := append([]string{fileDir}, dirs...)

	for _, ext := range TypeExt[f] {
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
