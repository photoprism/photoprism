package tz

import "strings"

// Sanitize returns the normalized name of a time zone a user or identity provider selected, or an empty
// string if its shape is invalid. It accepts Local, UTC, the offsets Name normalizes, and names made of
// letters, digits, "/", "_", "+" and "-", without checking whether the time zone database knows them.
func Sanitize(s string) string {
	s = strings.TrimSpace(s)

	switch {
	case s == "", len(s) > MaxLen:
		return ""
	case IsUTC(s):
		return UTC
	case IsLocal(s):
		return Local
	}

	if zone := NormalizeUtcOffset(s); zone != "" {
		return zone
	}

	if !validZoneName(s) {
		return ""
	}

	return s
}

// validZoneName reports whether s has the shape of an IANA time zone name such as America/Port-au-Prince.
// Offset names that NormalizeUtcOffset does not accept, such as UTC+99, are refused, except for the two
// IANA zones beyond UTC+12.
func validZoneName(s string) bool {
	if s == "Etc/GMT-13" || s == "Etc/GMT-14" {
		return true
	}

	for _, segment := range strings.Split(s, "/") {
		if segment == "" || !isLetter(segment[0]) {
			return false
		} else if upper := strings.ToUpper(segment); strings.HasPrefix(upper, UTC) || strings.HasPrefix(upper, GMT) {
			return false
		}

		for i := 0; i < len(segment); i++ {
			if b := segment[i]; !isLetter(b) && (b < '0' || b > '9') && b != '_' && b != '+' && b != '-' {
				return false
			}
		}
	}

	return true
}

// isLetter reports whether b is an ASCII letter.
func isLetter(b byte) bool {
	return b >= 'A' && b <= 'Z' || b >= 'a' && b <= 'z'
}
