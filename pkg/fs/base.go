package fs

import (
	"path/filepath"
	"strconv"
	"strings"
)

// StripExt removes the file type extension from a file name (if any).
func StripExt(name string) string {
	if end := strings.LastIndex(name, "."); end != -1 {
		name = name[:end]
	}

	return name
}

// StripKnownExt removes a known file type extension from a file name (if any).
func StripKnownExt(name string) string {
	if FileExt.Known(name) {
		name = StripExt(name)
	}

	return name
}

// Base returns the filename base without any extensions and path.
func Base(fileName string, stripSequence bool) string {
	name := StripKnownExt(StripExt(filepath.Base(fileName)))

	if !stripSequence {
		return name
	}

	// Strip numeric extensions like .00000, .00001, .4542353245,.... (at least 5 digits).
	if dot := strings.LastIndex(name, "."); dot != -1 && len(name[dot+1:]) >= 5 {
		if i, err := strconv.Atoi(name[dot+1:]); err == nil && i >= 0 {
			name = name[:dot]
		}
	}

	// Other common sequential naming schemes.
	if end := strings.Index(name, "("); end != -1 {
		// Copies created by Chrome & Windows, example: IMG_1234 (2).
		name = name[:end]
	} else if end := strings.Index(name, " copy"); end != -1 {
		// Copies created by OS X, example: IMG_1234 copy 2.
		name = name[:end]
	}

	name = strings.TrimSpace(name)

	return name
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
