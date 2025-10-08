package clean

import "strings"

// Thumb returns a sanitized thumbnail hash (40 hex characters) or an empty string when invalid.
func Thumb(s string) string {
	s = strings.TrimSpace(s)
	if len(s) != 40 {
		return ""
	}

	if h := Hex(s); len(h) == 40 {
		return h
	}

	return ""
}

// ThumbCrop returns a sanitized thumbnail crop hash (`hash-area`) or an empty string when invalid.
// Cropped thumbnails combine the base hash with a 12-hex crop descriptor separated by a dash.
func ThumbCrop(s string) string {
	s = strings.TrimSpace(s)
	if len(s) < 41 {
		return ""
	}

	if i := strings.IndexByte(s, '-'); i == -1 || i < 40 || i >= len(s)-1 {
		return ""
	} else {
		h := Thumb(s[:i])
		if h == "" {
			return ""
		}

		crop := Hex(s[i+1:])
		if len(crop) != 12 {
			return ""
		}

		return h + "-" + crop
	}
}
