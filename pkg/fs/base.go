package fs

import (
	"path/filepath"
	"regexp"
	"strings"
)

// StripSequenceRegex is the pattern used in StripSequence.
// Default pattern matches:
// - numeric extensions like .00000, .00001, .4542353245,.... (at least 5 digits) with: \.\d{5,}$
// - sequential naming schemes like IMG_1234 copy 2 with: copy.*$
// - sequential naming schemes like IMG_1234 (2) with: \(.*$
var StripSequenceRegex = regexp.MustCompile(`\.\d{5,}$| copy.*$|\(.*$`)

// StripSequence removes common sequence patterns at the end of file names.
func StripSequence(name string) string {
	idx := StripSequenceRegex.FindStringIndex(name)
	if idx != nil {
		name = name[0:idx[0]]
	}
	return strings.TrimSpace(name)
}

// BasePrefix returns the filename base without any extensions and path.
func BasePrefix(fileName string, stripSequence bool) string {
	name := StripKnownExt(StripExt(filepath.Base(fileName)))

	if !stripSequence {
		return name
	}

	return StripSequence(name)
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
	return filepath.Join(filepath.Dir(fileName), BasePrefix(fileName, stripSequence))
}
