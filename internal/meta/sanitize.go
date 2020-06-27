package meta

import (
	"regexp"
	"strings"
)

var DscTitleRegexp = regexp.MustCompile("\\D{3}[\\d_]\\d{4}(.JPG)?")

var UnwantedDescriptions = map[string]bool{
	"OLYMPUS DIGITAL CAMERA": true,
}

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

// SanitizeTitle normalizes titles and removes unwanted information.
func SanitizeTitle(value string) string {
	value = SanitizeString(value)

	if dsc := DscTitleRegexp.FindString(value); dsc == value {
		value = ""
	}

	return value
}

// SanitizeDescription normalizes descriptions and removes unwanted information.
func SanitizeDescription(value string) string {
	value = SanitizeString(value)

	if remove := UnwantedDescriptions[value]; remove {
		value = ""
	}

	return value
}
