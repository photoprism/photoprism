package meta

import (
	"math"
	"strconv"
	"strings"
)

const (
	// RatingMin is the lowest explicit star rating value.
	RatingMin = 0

	// RatingMax is the highest supported star rating value.
	RatingMax = 5
)

// SetRating stores a normalized star rating in metadata.
func (data *Data) SetRating(rating int) {
	if data == nil {
		return
	}

	if rating < RatingMin {
		rating = RatingMin
	} else if rating > RatingMax {
		rating = RatingMax
	}

	data.Rating = rating
	data.RatingSet = true
}

// ParseRating converts an EXIF or XMP rating value to a 0-5 star rating.
func ParseRating(s string) (rating int, ok bool) {
	s = strings.TrimSpace(s)

	if s == "" {
		return 0, false
	}

	value, err := strconv.ParseFloat(s, 64)

	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, false
	}

	if value != math.Trunc(value) {
		return 0, false
	}

	rating = int(value)

	if rating < RatingMin || rating > RatingMax {
		return 0, false
	}

	return rating, true
}

// ParseRatingPercent converts an EXIF RatingPercent value to a 0-5 star rating.
func ParseRatingPercent(s string) (rating int, ok bool) {
	s = strings.TrimSpace(s)

	if s == "" {
		return 0, false
	}

	value, err := strconv.ParseFloat(s, 64)

	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, false
	}

	if value < 0 || value > 100 {
		return 0, false
	}

	switch {
	case value == 0:
		return 0, true
	case value <= 12.5:
		return 1, true
	case value <= 37.5:
		return 2, true
	case value <= 62.5:
		return 3, true
	case value <= 87.5:
		return 4, true
	default:
		return 5, true
	}
}

// SetExiftoolRating stores the first supported EXIF/XMP star rating found in JSON metadata.
func (data *Data) SetExiftoolRating() {
	if data == nil || len(data.json) == 0 {
		return
	}

	for _, key := range []string{"Rating", "UserRating"} {
		if rating, ok := ParseRating(data.json[key]); ok {
			data.SetRating(rating)
			return
		}
	}

	if rating, ok := ParseRatingPercent(data.json["RatingPercent"]); ok {
		data.SetRating(rating)
	}
}
