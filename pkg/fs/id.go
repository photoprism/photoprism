package fs

import (
	"regexp"

	"github.com/photoprism/photoprism/pkg/rnd"
)

var DscNameRegexp = regexp.MustCompile("\\D{3}[\\d_]\\d{4}(.JPG)?")

// IsInt tests if the file base is an integer number.
func IsInt(base string) bool {
	if base == "" {
		return false
	}

	for _, r := range base {
		if r < 48 || r > 57 {
			return false
		}
	}

	return true
}

// IsID tests if the file name looks like an automatically created identifier.
func IsID(fileName string) bool {
	if fileName == "" {
		return false
	}

	base := Base(fileName, false)

	if IsHash(base) {
		return true
	}

	if IsInt(base) {
		return true
	}

	if dsc := DscNameRegexp.FindString(base); dsc == base {
		return true
	}

	if rnd.IsUID(base, 0) {
		return true
	}

	if IsCanonical(base) {
		return true
	}

	return false
}
