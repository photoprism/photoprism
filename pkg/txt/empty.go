package txt

import (
	"strings"
)

// Empty checks whether a string represents an empty, unset, or undefined value.
func Empty(s string) bool {
	if s == "" {
		return true
	} else if s = strings.Trim(s, "%* "); s == "" || s == "0" || s == "-1" || DateTimeDefault(s) {
		return true
	} else if s = strings.ToLower(s); s == "nil" || s == "null" || s == "none" || s == "nan" {
		return true
	}

	return false
}

// NotEmpty tests if a string does not represent an empty/invalid value.
func NotEmpty(s string) bool {
	return !Empty(s)
}

// EmptyDateTime tests if the string is empty or matches an unknown time pattern.
func EmptyDateTime(s string) bool {
	switch s {
	case "", "-", ":", "z", "Z", "nil", "null", "none", "nan", "NaN":
		return true
	case "0", "00", "0000", "0000:00:00", "00:00:00", "0000-00-00", "00-00-00":
		return true
	case "    :  :     :  :  ", "    -  -     -  -  ", "    -  -     :  :  ":
		// Exif default.
		return true
	case "0000:00:00 00:00:00", "0000-00-00 00-00-00", "0000-00-00 00:00:00":
		return true
	case "0001:01:01 00:00:00", "0001-01-01 00-00-00", "0001-01-01 00:00:00":
		// Go default.
		return true
	case "0001:01:01 00:00:00 +0000 UTC", "0001-01-01 00-00-00 +0000 UTC", "0001-01-01 00:00:00 +0000 UTC":
		// Go default with time zone.
		return true
	default:
		return false
	}
}

// DateTimeDefault tests if the datetime string is not empty and not a default value.
func DateTimeDefault(s string) bool {
	switch s {
	case "1970-01-01", "1970-01-01 00:00:00", "1970:01:01 00:00:00":
		// Unix epoch.
		return true
	case "1980-01-01", "1980-01-01 00:00:00", "1980:01:01 00:00:00":
		// Windows default.
		return true
	case "2002-12-08 12:00:00", "2002:12:08 12:00:00":
		// Android Bug: https://issuetracker.google.com/issues/36967504
		return true
	default:
		return EmptyDateTime(s)
	}
}
