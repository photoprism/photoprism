package fs

import (
	"os"
	"strings"
)

// RelativeName returns the file name relative to directory.
func RelativeName(fileName, directory string) string {
	if index := strings.Index(fileName, directory); index == 0 {
		if index := strings.LastIndex(directory, string(os.PathSeparator)); index == len(directory)-1 {
			pos := len(directory)
			return fileName[pos:]
		} else if index := strings.LastIndex(directory, string(os.PathSeparator)); index != len(directory) {
			pos := len(directory) + 1
			return fileName[pos:]
		}
	}

	return fileName
}
