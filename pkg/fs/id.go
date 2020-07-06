package fs

import (
	"regexp"

	"github.com/photoprism/photoprism/pkg/rnd"
)

var DscNameRegexp = regexp.MustCompile("\\D{3}[\\d_]\\d{4}(.JPG)?")

// IsInt tests if the file base is an integer number.
func IsInt(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if r < 48 || r > 57 {
			return false
		}
	}

	return true
}

// IsAsciiID tests if the string is a file name that only contains uppercase ascii letters and numbers like "IQVG4929".
func IsAsciiID(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if (r < 65 || r > 90) && (r < 48 || r > 57) {
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

	if IsAsciiID(base) {
		return true
	}

	return false
}
