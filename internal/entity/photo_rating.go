package entity

const (
	// PhotoRatingUnknown marks photos without an imported or manually assigned rating.
	PhotoRatingUnknown = -1

	// PhotoRatingMin is the lowest explicit star rating value.
	PhotoRatingMin = 0

	// PhotoRatingMax is the highest supported star rating value.
	PhotoRatingMax = 5
)

// HasRating reports whether the photo has an explicit star rating value.
func (m *Photo) HasRating() bool {
	return m != nil && m.PhotoRating >= PhotoRatingMin && m.PhotoRating <= PhotoRatingMax
}

// SetRating updates the photo star rating when the supplied source has enough priority.
func (m *Photo) SetRating(rating int, source string) bool {
	if m == nil || source == "" {
		return false
	}

	if rating < PhotoRatingMin || rating > PhotoRatingMax {
		return false
	}

	if (SrcPriority[source] < SrcPriority[m.RatingSrc]) && m.HasRating() {
		return false
	}

	newRating := int8(rating) //nolint:gosec // rating is validated above.

	if m.PhotoRating == newRating && m.RatingSrc == source {
		return false
	}

	m.PhotoRating = newRating
	m.RatingSrc = source

	return true
}
