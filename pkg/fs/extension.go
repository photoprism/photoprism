package fs

import (
	"path/filepath"
	"strings"
)

// Ext returns all extension of a file name including the dots.
func Ext(name string) string {
	ext := filepath.Ext(name)
	name = StripExt(name)

	if Extensions.Known(name) {
		ext = filepath.Ext(name) + ext
	}

	return ext
}

// StripExt removes the file type extension from a file name (if any).
func StripExt(name string) string {
	if end := strings.LastIndex(name, "."); end != -1 {
		name = name[:end]
	}

	return name
}

// StripKnownExt removes all known file type extension from a file name (if any).
func StripKnownExt(name string) string {
	for Extensions.Known(name) {
		name = StripExt(name)
	}

	return name
}
