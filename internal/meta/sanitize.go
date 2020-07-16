package meta

import (
	"regexp"
	"strings"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

var UnwantedDescriptions = map[string]bool{
	"OLYMPUS DIGITAL CAMERA": true, // Olympus
	"SAMSUNG":                true, // Samsung
	"rhdr":                   true, // Huawei
	"hdrpl":                  true,
	"fbt":                    true,
	"mon":                    true,
	"nor":                    true,
	"dav":                    true,
	"mde":                    true,
	"mde_soft":               true,
	"edf":                    true,
	"btfmdn":                 true,
	"btf":                    true,
	"btfhdr":                 true,
	"frem":                   true,
	"oznor":                  true,
	"rpt":                    true,
	"burst":                  true,
	"sdr_HDRB":               true,
}

var LowerCaseRegexp = regexp.MustCompile("[a-z0-9_\\-]+")

// SanitizeString removes unwanted character from an exif value string.
func SanitizeString(s string) string {
	if s == "" {
		return ""
	}

	s = strings.TrimSpace(s)

	return strings.Replace(s, "\"", "", -1)
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
func SanitizeTitle(title string) string {
	s := SanitizeString(title)

	if s == "" {
		return ""
	}

	result := fs.StripKnownExt(s)

	if fs.IsGenerated(result) || txt.IsTime(result) {
		result = ""
	} else if result == s {
		// Do nothing.
	} else if found := LowerCaseRegexp.FindString(result); found != result {
		result = strings.ReplaceAll(strings.ReplaceAll(result, "_", " "), "  ", " ")
	} else if formatted := txt.FileTitle(s); formatted != "" {
		result = formatted
	} else {
		result = txt.Title(strings.ReplaceAll(result, "-", " "))
	}

	return result
}

// SanitizeDescription normalizes descriptions and removes unwanted information.
func SanitizeDescription(s string) string {
	s = SanitizeString(s)

	if s == "" {
		return ""
	} else if remove := UnwantedDescriptions[s]; remove {
		s = ""
	} else if strings.HasPrefix(s, "DCIM\\") && !strings.Contains(s, " ") {
		s = ""
	}

	return s
}
