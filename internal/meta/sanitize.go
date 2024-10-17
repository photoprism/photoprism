package meta

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

var UnwantedDescriptions = map[string]bool{
	"Created by Imlib":         true, // Apps
	"iClarified":               true,
	"OLYMPUS DIGITAL CAMERA":   true, // Olympus
	"SAMSUNG":                  true, // Samsung
	"SAMSUNG CAMERA PICTURES":  true,
	"<Digimax i5, Samsung #1>": true,
	"SONY DSC":                 true, // Sony
	"Scanner":                  true, // KODAK Slide N Scan
	"rhdr":                     true, // Huawei
	"hdrpl":                    true,
	"oznorWO":                  true,
	"frontbhdp":                true,
	"fbt":                      true,
	"rbt":                      true,
	"ptr":                      true,
	"fbthdr":                   true,
	"btr":                      true,
	"mon":                      true,
	"nor":                      true,
	"dav":                      true,
	"mde":                      true,
	"mde_soft":                 true,
	"edf":                      true,
	"btfmdn":                   true,
	"btf":                      true,
	"btfhdr":                   true,
	"frem":                     true,
	"oznor":                    true,
	"rpt":                      true,
	"burst":                    true,
	"sdr_HDRB":                 true,
	"cof":                      true,
	"qrf":                      true,
	"fshbty":                   true,
	"binary comment":           true, // Other
	"default":                  true,
	"Exif_JPEG_PICTURE":        true,
	"DVC 10.1 HDMI":            true,
	"charset=Ascii":            true,
}

var LowerCaseRegexp = regexp.MustCompile("[a-z\\d_\\-]+")

// SanitizeUnicode returns the string as valid Unicode with whitespace trimmed.
func SanitizeUnicode(s string) string {
	if s == "" {
		return ""
	}

	return clean.Unicode(strings.TrimSpace(s))
}

// SanitizeString removes unwanted character from an exif value string.
func SanitizeString(s string) string {
	if s == "" {
		return ""
	}

	if strings.HasPrefix(s, "string with binary data") {
		return ""
	} else if strings.HasPrefix(s, "(Binary data") {
		return ""
	}

	return SanitizeUnicode(strings.Replace(s, "\"", "", -1))
}

// SanitizeUID normalizes unique IDs found in XMP or Exif metadata.
func SanitizeUID(s string) string {
	s = SanitizeString(s)

	if len(s) < 15 {
		return ""
	}

	if start := strings.LastIndex(s, ":"); start != -1 {
		s = s[start+1:]
	}

	// Not a unique ID?
	if len(s) < 15 || len(s) > 36 {
		s = ""
	}

	return strings.ToLower(s)
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

	switch {
	case s == "":
		return ""
	case UnwantedDescriptions[s]:
		return ""
	case strings.HasPrefix(s, "DCIM\\") && !strings.Contains(s, " "):
		return ""
	default:
		return s
	}
}

// SanitizeMeta normalizes metadata fields that may contain JSON arrays like keywords and subject.
func SanitizeMeta(s string) string {
	if s == "" {
		return ""
	}

	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		var words []string

		if err := json.Unmarshal([]byte(s), &words); err != nil {
			return s
		}

		s = strings.Join(words, ", ")
	} else {
		s = SanitizeString(s)
	}

	return s
}
