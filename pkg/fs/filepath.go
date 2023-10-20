package fs

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// RelatedMediaFileSuffix is a regular expression that matches suffixes of related media files,
// see https://github.com/photoprism/photoprism/issues/2983 (Support Live Photos downloaded with "iCloudPD").
var RelatedMediaFileSuffix = regexp.MustCompile(`(?i)_(jpg|jpeg|hevc)$`)

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

// BasePrefix returns the filename base without any extensions and path.
func BasePrefix(fileName string, stripSequence bool) string {
	fileBase := StripKnownExt(StripExt(filepath.Base(fileName)))

	if !stripSequence {
		return fileBase
	}

	return StripSequence(fileBase)
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

// RelatedFilePathPrefix returns the absolute file path and name prefix without file extensions and media file
// suffixes to be ignored for comparison, see https://github.com/photoprism/photoprism/issues/2983.
func RelatedFilePathPrefix(fileName string, stripSequence bool) string {
	if fileName == "" {
		return ""
	}

	return RelatedMediaFileSuffix.ReplaceAllString(AbsPrefix(fileName, stripSequence), "")
}
