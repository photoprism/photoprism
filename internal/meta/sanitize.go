package meta

import (
	"strings"
)

// SanitizeString removes unwanted character from an exif value string.
func SanitizeString(value string) string {
	value = strings.TrimSpace(value)
	return strings.Replace(value, "\"", "", -1)
}

// SanitizeUID normalizes unique IDs found in XMP or Exif metadata.
func SanitizeUID(value string) string {
	value = SanitizeString(value)

	if start := strings.LastIndex(value, ":"); start != -1 {
		value = value[start+1:]
	}

	// Not a unique ID?
	if len(value) < 15 || len(value) > 36 {
		value = ""
	}

	return strings.ToLower(value)
}
