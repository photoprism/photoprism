package fs

import (
	"path/filepath"
	"strconv"
	"strings"
)

// StripSequence removes common sequence patterns at the end of file names.
func StripSequence(fileName string) string {
	if fileName == "" {
		return ""
	}

	// Strip numeric extensions like .00000, .00001, .4542353245,.... (at least 5 digits).
	if dot := strings.LastIndex(fileName, "."); dot != -1 && len(fileName[dot+1:]) >= 5 {
		if i, err := strconv.Atoi(fileName[dot+1:]); err == nil && i >= 0 {
			fileName = fileName[:dot]
		}
	}

	// Other common sequential naming schemes.
	if end := strings.Index(fileName, "("); end != -1 {
		// Copies created by Chrome & Windows, example: IMG_1234 (2).
		fileName = fileName[:end]
	} else if end := strings.Index(fileName, " copy"); end != -1 {
		// Copies created by OS X, example: IMG_1234 copy 2.
		fileName = fileName[:end]
	}

	fileName = strings.TrimSpace(fileName)

	return fileName
}

// BasePrefix returns the filename base without any extensions and path. A sequence is only stripped
// if a name remains, e.g. "(1).jpg" keeps "(1)", and a name whose base is a dot name keeps its
// extensions, e.g. "..jpg", so the prefix of a file name never names a folder.
func BasePrefix(fileName string, stripSequence bool) string {
	baseName := filepath.Base(fileName)
	fileBase := StripKnownExt(StripExt(baseName))

	if NoBaseName(fileBase) {
		if NoBaseName(baseName) {
			return ""
		}

		fileBase = baseName
	}

	if !stripSequence {
		return fileBase
	} else if prefix := StripSequence(fileBase); !NoBaseName(prefix) {
		return prefix
	}

	return fileBase
}

// NoBaseName reports whether a base name is empty or a dot name, which names a folder rather than a file
// when joined with a path.
func NoBaseName(base string) bool {
	return base == "" || base == "." || base == ".."
}

// RelPrefix returns the relative filename.
func RelPrefix(fileName, dir string, stripSequence bool) string {
	if name := RelName(fileName, dir); name != "" {
		return AbsPrefix(name, stripSequence)
	}

	return BasePrefix(fileName, stripSequence)
}

// AbsPrefix returns the directory and base filename without any extensions.
func AbsPrefix(fileName string, stripSequence bool) string {
	if fileName == "" {
		return ""
	}

	return filepath.Join(filepath.Dir(fileName), BasePrefix(fileName, stripSequence))
}
