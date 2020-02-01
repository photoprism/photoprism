package photoprism

import (
	"strings"
	"time"
)

// NonCanonical returns true if the file basename is not canonical.
func NonCanonical(basename string) bool {
	if len(basename) != 28 {
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

// CanonicalName returns a canonical name based on time and hash.
func CanonicalName(date time.Time, hash string) string {
	var postfix string

	if len(hash) > 12 {
		postfix = strings.ToUpper(hash[:12])
	} else {
		postfix = "NOTFOUND"
	}

	return date.Format("20060102_150405_") + postfix
}
