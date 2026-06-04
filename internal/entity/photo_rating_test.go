package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhoto_SetRating(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		photo := NewPhoto(true)
		photo.PhotoUID = "p123456789"
		photo.PhotoRating = 0

		err := photo.SetRating(3)
		assert.NoError(t, err)
		assert.Equal(t, int8(3), photo.PhotoRating)
	})
	t.Run("Unrated", func(t *testing.T) {
		photo := NewPhoto(true)
		photo.PhotoUID = "p234567890"
		photo.PhotoRating = 3

		err := photo.SetRating(0)
		assert.NoError(t, err)
		assert.Equal(t, int8(0), photo.PhotoRating)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		photo := NewPhoto(true)
		photo.PhotoUID = "p345678901"

		err := photo.SetRating(6)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rating must be between 0 and 5")
	})
	t.Run("Negative", func(t *testing.T) {
		photo := NewPhoto(true)
		photo.PhotoUID = "p456789012"

		err := photo.SetRating(-1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rating must be between 0 and 5")
	})
}
