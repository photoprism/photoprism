package meta

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRating(t *testing.T) {
	for value, expected := range map[string]int{
		"0":   0,
		"1":   1,
		"5":   5,
		" 4 ": 4,
	} {
		rating, ok := ParseRating(value)

		assert.True(t, ok, value)
		assert.Equal(t, expected, rating, value)
	}

	for _, value := range []string{"", "-1", "6", "2.5", "nan", "abc"} {
		_, ok := ParseRating(value)

		assert.False(t, ok, value)
	}
}

func TestParseRatingPercent(t *testing.T) {
	for value, expected := range map[string]int{
		"0":     0,
		"1":     1,
		"25":    2,
		"50":    3,
		"75":    4,
		"99.99": 5,
		"100":   5,
	} {
		rating, ok := ParseRatingPercent(value)

		assert.True(t, ok, value)
		assert.Equal(t, expected, rating, value)
	}

	for _, value := range []string{"", "-0.1", "100.1", "nan", "abc"} {
		_, ok := ParseRatingPercent(value)

		assert.False(t, ok, value)
	}
}

func TestData_SetExiftoolRating(t *testing.T) {
	t.Run("Rating", func(t *testing.T) {
		data := Data{json: map[string]string{"Rating": "0"}}

		data.SetExiftoolRating()

		assert.True(t, data.RatingSet)
		assert.Equal(t, 0, data.Rating)
	})

	t.Run("UserRating", func(t *testing.T) {
		data := Data{json: map[string]string{"UserRating": "5"}}

		data.SetExiftoolRating()

		assert.True(t, data.RatingSet)
		assert.Equal(t, 5, data.Rating)
	})

	t.Run("RatingPercent", func(t *testing.T) {
		data := Data{json: map[string]string{"RatingPercent": "50"}}

		data.SetExiftoolRating()

		assert.True(t, data.RatingSet)
		assert.Equal(t, 3, data.Rating)
	})

	t.Run("Invalid", func(t *testing.T) {
		data := Data{json: map[string]string{"Rating": "6", "UserRating": "-1", "RatingPercent": "abc"}}

		data.SetExiftoolRating()

		assert.False(t, data.RatingSet)
		assert.Equal(t, 0, data.Rating)
	})
}
