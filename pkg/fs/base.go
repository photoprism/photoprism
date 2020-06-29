package fs

import (
	"path/filepath"
	"strconv"
	"strings"
)

// Base returns the filename base without any extensions and path.
func Base(fileName string, stripSequence bool) string {
	basename := filepath.Base(fileName)

	// Strip file type extension.
	if end := strings.LastIndex(basename, "."); end != -1 {
		basename = basename[:end]
	}

	if !stripSequence {
		return basename
	}

	// Strip numeric extensions like .00000, .00001, .4542353245,.... (at least 5 digits).
	if dot := strings.LastIndex(basename, "."); dot != -1 && len(basename[dot+1:]) >= 5 {
		if i, err := strconv.Atoi(basename[dot+1:]); err == nil && i >= 0 {
			basename = basename[:dot]
		}
	}

	// Other common sequential naming schemes.
	if end := strings.Index(basename, "("); end != -1 {
		// Copies created by Chrome & Windows, example: IMG_1234 (2).
		basename = basename[:end]
	} else if end := strings.Index(basename, " copy"); end != -1 {
		// Copies created by OS X, example: IMG_1234 copy 2.
		basename = basename[:end]
	}

	basename = strings.TrimSpace(basename)

	return basename
}

// RelBase returns the relative filename.
func RelBase(fileName, dir string, stripSequence bool) string {
	if name := Rel(fileName, dir); name != "" {
		return AbsBase(name, stripSequence)
	}

	return Base(fileName, stripSequence)
}

// AbsBase returns the directory and base filename without any extensions.
func AbsBase(fileName string, stripSequence bool) string {
	return filepath.Join(filepath.Dir(fileName), Base(fileName, stripSequence))
}
