package meta

import (
	"math"
	"strconv"
	"strings"

	"github.com/photoprism/photoprism/pkg/txt"
)

// normalizeRating clamps metadata star ratings to PhotoPrism's supported range.
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

// normalizeRatingPercent converts a percent-based rating to the nearest 0-5 star value.
func normalizeRatingPercent(percent int) int {
	return normalizeRating((percent + 10) / 20)
}

// parseRatingValue parses a metadata rating and reports if a value was present.
func parseRatingValue(s string) (int, bool) {
	s = strings.TrimSpace(s)

	if s == "" {
		return 0, false
	}

	f, err := strconv.ParseFloat(txt.Numeric(s), 64)

	if err != nil {
		return 0, false
	}

	if f < 0 {
		return 0, false
	}

	return normalizeRating(int(math.Round(f))), true
}

// parseRatingPercent parses a percent metadata rating and reports if a value was present.
func parseRatingPercent(s string) (int, bool) {
	s = strings.TrimSpace(s)

	if s == "" {
		return 0, false
	}

	i, err := strconv.Atoi(txt.Numeric(s))

	if err != nil {
		return 0, false
	}

	if i < 0 {
		return 0, false
	}

	return normalizeRatingPercent(i), true
}
