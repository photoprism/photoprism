package meta

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/media/projection"

	"github.com/photoprism/photoprism/pkg/txt"
)

const (
	KeywordFlash           = "flash"
	KeywordHdr             = "hdr"
	KeywordBurst           = "burst"
	KeywordPanorama        = "panorama"
	KeywordEquirectangular = string(projection.Equirectangular)
)

// Keywords represents a list of metadata keywords.
type Keywords []string

// String returns a string containing all keywords.
func (w Keywords) String() string {
	return strings.Join(w, ", ")
}

var AutoKeywords = []string{KeywordHdr, KeywordBurst, KeywordPanorama, KeywordEquirectangular}

// AddKeywords appends keywords.
func (data *Data) AddKeywords(w string) {
	w = strings.ToLower(SanitizeMeta(w))

	if len(w) < 1 {
		return
	}

	data.Keywords = txt.AddToWords(data.Keywords, w)
}

// AutoAddKeywords automatically appends relevant keywords from a string (e.g. description).
func (data *Data) AutoAddKeywords(s string) {
	s = strings.ToLower(SanitizeMeta(s))

	if len(s) < 1 {
		return
	}

	for _, w := range AutoKeywords {
		if strings.Contains(s, w) {
			data.AddKeywords(w)
			if w == KeywordHdr {
				data.ImageType = ImageTypeHDR
			}
		}
	}
}
