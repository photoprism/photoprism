package meta

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeRating(t *testing.T) {
	assert.Equal(t, 0, normalizeRating(-1))
	assert.Equal(t, 0, normalizeRating(0))
	assert.Equal(t, 3, normalizeRating(3))
	assert.Equal(t, 5, normalizeRating(7))
}

func TestNormalizeRatingPercent(t *testing.T) {
	assert.Equal(t, 0, normalizeRatingPercent(0))
	assert.Equal(t, 1, normalizeRatingPercent(10))
	assert.Equal(t, 4, normalizeRatingPercent(75))
	assert.Equal(t, 5, normalizeRatingPercent(100))
}

func TestParseRatingValue(t *testing.T) {
	rating, ok := parseRatingValue("4")
	assert.True(t, ok)
	assert.Equal(t, 4, rating)

	rating, ok = parseRatingValue("4.6")
	assert.True(t, ok)
	assert.Equal(t, 5, rating)

	rating, ok = parseRatingValue("0")
	assert.True(t, ok)
	assert.Equal(t, 0, rating)

	rating, ok = parseRatingValue("-1")
	assert.False(t, ok)
	assert.Equal(t, 0, rating)

	rating, ok = parseRatingValue("")
	assert.False(t, ok)
	assert.Equal(t, 0, rating)
}

func TestParseRatingPercent(t *testing.T) {
	rating, ok := parseRatingPercent("75")
	assert.True(t, ok)
	assert.Equal(t, 4, rating)

	rating, ok = parseRatingPercent("")
	assert.False(t, ok)
	assert.Equal(t, 0, rating)

	rating, ok = parseRatingPercent("-1")
	assert.False(t, ok)
	assert.Equal(t, 0, rating)
}
