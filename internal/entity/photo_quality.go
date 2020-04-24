package entity

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/txt"
)

var QualityBlacklist = map[string]bool{
	"screenshot":  true,
	"screenshots": true,
	"info":        true,
}

// QualityScore returns a score based on photo properties like size and metadata.
func (m *Photo) QualityScore() (score int) {
	if m.PhotoFavorite {
		score += 3
	}

	if m.TakenSrc != SrcAuto {
		score++
	}

	if m.HasLatLng() {
		score++
	}

	if m.PhotoResolution >= 2 {
		score++
	}

	blacklisted := false

	if m.Description.PhotoKeywords != "" {
		keywords := txt.Words(m.Description.PhotoKeywords)

		for _, w := range keywords {
			w = strings.ToLower(w)

			if _, ok := QualityBlacklist[w]; ok {
				blacklisted = true
				break
			}
		}
	}

	if !blacklisted {
		score++
	}

	return score
}
