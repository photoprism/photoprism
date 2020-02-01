package photoprism

import (
	"strings"
	"time"
)

// NonCanonical returns true if the file basename is not canonical.
func NonCanonical(basename string) bool {
	if len(basename) != 24 {
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

// CanonicalName returns a canonical name based on time and CRC32 checksum.
func CanonicalName(date time.Time, checksum string) string {
	if len(checksum) != 8 {
		checksum = "ERROR000"
	} else {
		checksum = strings.ToUpper(checksum)
	}

	return date.Format("20060102_150405_") + checksum
}
