package entity

import (
	"strings"
	"time"

	"github.com/photoprism/photoprism/pkg/txt"
)

var NonPhotographicKeywords = map[string]bool{
	"screenshot":  true,
	"screenshots": true,
	"info":        true,
}

var (
	year2008 = time.Date(2008, 1, 1, 0, 0, 0, 0, time.UTC)
	year2012 = time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC)
)

// QualityScore returns a score based on photo properties like size and metadata.
func (m *Photo) QualityScore() (score int) {
	if m.PhotoFavorite {
		score += 3
	}

	if SrcPriority[m.TakenSrc] > SrcPriority[SrcEstimate] {
		score++
	}

	if m.TrustedLocation() {
		score++
	}

	if m.TakenAt.Before(year2008) {
		score++
	} else if m.TakenAt.Before(year2012) && m.PhotoResolution >= 1 {
		score++
	} else if m.PhotoResolution >= 2 {
		score++
	}

	if !m.IsNonPhotographic() {
		score++
	}

	if score < 3 && (m.PhotoType != MediaImage || m.EditedAt != nil) {
		score = 3
	}

	return score
}

// UpdateQuality updates the photo quality attribute.
func (m *Photo) UpdateQuality() error {
	if m.DeletedAt != nil || m.PhotoQuality < 0 {
		return nil
	}

	m.PhotoQuality = m.QualityScore()

	return m.Update("PhotoQuality", m.PhotoQuality)
}

// IsNonPhotographic checks whether the image appears to be non-photographic.
func (m *Photo) IsNonPhotographic() (result bool) {
	if m.PhotoType == MediaUnknown || m.PhotoType == MediaVector || m.PhotoType == MediaAnimated {
		return true
	}

	details := m.GetDetails()

	if details.Keywords != "" {
		keywords := txt.Words(details.Keywords)

		for _, w := range keywords {
			w = strings.ToLower(w)

			if _, ok := NonPhotographicKeywords[w]; ok {
				result = true
				break
			}
		}
	}

	return result
}
