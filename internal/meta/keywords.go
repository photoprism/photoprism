package meta

import "strings"

const (
	KeywordFlash           = "flash"
	KeywordHdr             = "hdr"
	KeywordBurst           = "burst"
	KeywordPanorama        = "panorama"
	KeywordEquirectangular = "equirectangular"
)

var AutoKeywords = []string{KeywordHdr, KeywordBurst, KeywordPanorama, KeywordEquirectangular}

// AddKeyword appends a keyword if not exists.
func (data *Data) AddKeyword(w string) {
	w = strings.ToLower(SanitizeString(w))

	if len(w) < 3 {
		return
	}

	if !strings.Contains(data.Keywords, w) {
		if data.Keywords == "" {
			data.Keywords = w
		} else {
			data.Keywords += ", " + w
		}
	}
}

// AutoAddKeywords automatically adds relevant keywords from a string (e.g. description).
func (data *Data) AutoAddKeywords(s string) {
	s = strings.ToLower(SanitizeString(s))

	if len(s) < 3 {
		return
	}

	for _, w := range AutoKeywords {
		if strings.Contains(s, w) {
			data.AddKeyword(w)
		}
	}
}
