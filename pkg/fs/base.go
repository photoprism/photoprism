package fs

import (
	"path/filepath"
	"strings"
)

// Base returns the filename base without any extensions and path.
func Base(fileName string, stripSequence bool) string {
	basename := filepath.Base(fileName)

	if end := strings.Index(basename, "."); end != -1 {
		// ignore everything behind the first dot in the file name
		basename = basename[:end]
	}

	if !stripSequence {
		return basename
	}

	// common sequential naming schemes
	if end := strings.Index(basename, " ("); end != -1 {
		// copies created by Chrome & Windows, example: IMG_1234 (2)
		basename = basename[:end]
	} else if end := strings.Index(basename, " copy"); end != -1 {
		// copies created by OS X, example: IMG_1234 copy 2
		basename = basename[:end]
	}

	return basename
}

// AbsBase returns the directory and base filename without any extensions.
func AbsBase(fileName string, stripSequence bool) string {
	return filepath.Join(filepath.Dir(fileName), Base(fileName, stripSequence))
}
