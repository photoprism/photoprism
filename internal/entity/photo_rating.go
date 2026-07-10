package entity

// Star rating range: 0 means unrated, 1 to 5 stars otherwise. RatingSrc
// distinguishes photos that have never been rated (empty source) from photos
// that were explicitly rated 0 stars (non-empty source), so both states
// remain queryable without a nullable column.
const (
	RatingMin = 0
	RatingMax = 5
)

// HasRating reports whether a rating has been assigned from any source,
// including an explicit rating of 0 stars.
func (m *Photo) HasRating() bool {
	return m.RatingSrc != ""
}

// GetRating returns the star rating from 0 to 5.
func (m *Photo) GetRating() int {
	return m.PhotoRating
}

// GetRatingSrc returns the rating source, if any.
func (m *Photo) GetRatingSrc() string {
	return m.RatingSrc
}

// SetRating stores the star rating when the value is within range and the
// source priority is sufficient, following the same rules as SetCaption.
func (m *Photo) SetRating(rating int, source string) {
	if rating < RatingMin || rating > RatingMax {
		return
	}

	if (SrcPriority[source] < SrcPriority[m.RatingSrc]) && m.HasRating() {
		return
	}

	m.PhotoRating = rating
	m.RatingSrc = source
}
