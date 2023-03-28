package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileName returns the relative filename with the same base and a given extension in a directory.
func FileName(fileName, dirName, baseDir, fileExt string) string {
	fileDir := filepath.Dir(fileName)

	if dirName == "" || dirName == "." {
		dirName = fileDir
	} else if fileDir != dirName {
		if filepath.IsAbs(dirName) {
			dirName = filepath.Join(dirName, RelName(fileDir, baseDir))
		} else {
			dirName = filepath.Join(fileDir, dirName)
		}
	}

	if err := os.MkdirAll(dirName, ModeDir); err != nil {
		fmt.Println(err.Error())
		return ""
	}

	result := filepath.Join(dirName, filepath.Base(fileName)) + fileExt

	return result
}

// RelName returns the file name relative to a directory.
func RelName(fileName, dir string) string {
	if fileName == dir {
		return ""
	}

	if dir == "" {
		return fileName
	}

	if index := strings.Index(fileName, dir); index == 0 {
		if index := strings.LastIndex(dir, string(os.PathSeparator)); index == len(dir)-1 {
			pos := len(dir)
			return fileName[pos:]
		} else if index := strings.LastIndex(dir, string(os.PathSeparator)); index != len(dir) {
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
