package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Base returns the filename base without any extensions and path.
func Base(fileName string, stripSequence bool) string {
	basename := filepath.Base(fileName)

	// strip file type extension
	if end := strings.LastIndex(basename, "."); end != -1 {
		basename = basename[:end]
	}

	if !stripSequence {
		return basename
	}

	// strip numeric extensions like .0000, .0001, .4542353245,....
	if dot := strings.LastIndex(basename, "."); dot != -1 && len(basename[dot+1:]) >= 5 {
		if i, err := strconv.Atoi(basename[dot+1:]); err == nil && i >= 0 {
			basename = basename[:dot]
		}
	}

	// other common sequential naming schemes
	if end := strings.Index(basename, "("); end != -1 {
		// copies created by Chrome & Windows, example: IMG_1234 (2)
		basename = basename[:end]
	} else if end := strings.Index(basename, " copy"); end != -1 {
		// copies created by OS X, example: IMG_1234 copy 2
		basename = basename[:end]
	}

	basename = strings.TrimSpace(basename)

	return basename
}

// RelativeBase returns the relative filename.
func RelativeBase(fileName, dir string, stripSequence bool) string {
	if name := RelativeName(fileName, dir); name != "" {
		return AbsBase(name, stripSequence)
	}

	return Base(fileName, stripSequence)
}

// AbsBase returns the directory and base filename without any extensions.
func AbsBase(fileName string, stripSequence bool) string {
	return filepath.Join(filepath.Dir(fileName), Base(fileName, stripSequence))
}

// SubFileName returns the a filename with the same base name and a given extension in a sub directory.
func SubFileName(fileName, subDir, fileExt string, stripSequence bool) string {
	baseName := Base(fileName, stripSequence)
	dirName := filepath.Join(filepath.Dir(fileName), subDir)

	if err := os.MkdirAll(dirName, os.ModePerm); err != nil {
		fmt.Println(err.Error())
		return ""
	}

	result := filepath.Join(dirName, baseName) + fileExt

	return result
}
