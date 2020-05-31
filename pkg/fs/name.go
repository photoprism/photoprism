package fs

import (
	"os"
	"strings"
)

// RelativeName returns the file name relative to directory.
func RelativeName(fileName, dir string) string {
	if fileName == dir {
		return ""
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
