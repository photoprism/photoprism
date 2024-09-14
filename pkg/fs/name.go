package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileName returns the file path for a sidecar file with the specified extension.
func FileName(fileName, dirName, baseDir, fileExt string) (string, error) {
	if fileName == "" {
		return "", fmt.Errorf("file name is empty")
	} else if fileExt == "" {
		return "", fmt.Errorf("file extension is empty")
	}

	dir := filepath.Dir(fileName)

	if dirName == "" || dirName == "." {
		dirName = dir
	} else if dir != dirName {
		if filepath.IsAbs(dirName) {
			dirName = filepath.Join(dirName, RelName(dir, baseDir))
		} else {
			dirName = filepath.Join(dir, dirName)
		}
	}

	// Create parent directories if they do not exist yet.
	if err := MkdirAll(dirName); err != nil {
		return "", err
	}

	// Compose and return file path.
	result := filepath.Join(dirName, filepath.Base(fileName)) + fileExt

	return result, nil
}

// RelName returns the file name relative to a directory.
func RelName(fileName, dir string) string {
	if fileName == dir {
		return ""
	}

	if dir == "" {
		return fileName
	}

	if i := strings.Index(fileName, dir); i == 0 {
		if i = strings.LastIndex(dir, string(os.PathSeparator)); i == len(dir)-1 {
			pos := len(dir)
			return fileName[pos:]
		} else if i = strings.LastIndex(dir, string(os.PathSeparator)); i != len(dir) {
			pos := len(dir) + 1
			return fileName[pos:]
		}
	}

	return fileName
}

// FileNameHidden tests is a file name belongs to a hidden file.
func FileNameHidden(name string) bool {
	if len(name) == 0 {
		return false
	}

	name = filepath.Base(name)

	// Hidden files and folders starting with "." or "@" should be ignored.
	switch name[0:1] {
	case ".", "@":
		return true
	}

	if len(name) == 1 {
		return false
	}

	// File paths starting with _. and __ like __MACOSX should be ignored.
	switch name[0:2] {
	case "_.", "__":
		return true
	}

	return false
}
