package fs

import (
	"strings"
	"time"
)

// NonCanonical returns true if the file basename is NOT canonical.
func NonCanonical(basename string) bool {
	if l := len(basename); l != 22 && l != 24 {
		return true
	}

	if strings.Count(basename, "_") != 2 {
		return true
	}

	if strings.ContainsAny(basename, "-~!@#$%^&*()+=<>?,.") {
		return true
	}

	return false
}

// IsCanonical returns true if the file basename is canonical.
func IsCanonical(basename string) bool {
	return !NonCanonical(basename)
}

// CanonicalName returns a canonical name based on time and CRC32 checksum.
func CanonicalName(date time.Time, checksum, pattern string) string {
	if len(checksum) != 8 {
		checksum = "EEEEEEEE"
	} else {
		checksum = strings.ToUpper(checksum)
	}

	if pattern == "" {
		pattern = "20060102_150405_"
	}

	return date.Format(pattern) + checksum
}
