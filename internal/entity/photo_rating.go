package entity

// normalizeRating clamps user star ratings to PhotoPrism's supported range.
func normalizeRating(rating int) int {
	switch {
	case rating < 0:
		return 0
	case rating > 5:
		return 5
	default:
		return rating
	}
}

// SetRating updates the photo star rating if the metadata source has sufficient priority.
func (m *Photo) SetRating(rating int, source string) bool {
	if m == nil {
		return false
	}

	if source == "" {
		source = SrcAuto
	}

	rating = normalizeRating(rating)

	if SrcPriority[source] < SrcPriority[m.RatingSrc] {
		return false
	}

	if m.PhotoRating == rating && m.RatingSrc == source {
		return false
	}

	m.PhotoRating = rating
	m.RatingSrc = source

	return true
}
