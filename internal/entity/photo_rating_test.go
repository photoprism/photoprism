package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeRating(t *testing.T) {
	assert.Equal(t, 0, normalizeRating(-1))
	assert.Equal(t, 0, normalizeRating(0))
	assert.Equal(t, 4, normalizeRating(4))
	assert.Equal(t, 5, normalizeRating(8))
}

func TestPhoto_SetRating(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		var photo *Photo
		assert.False(t, photo.SetRating(4, SrcManual))
	})

	t.Run("Set", func(t *testing.T) {
		photo := &Photo{}
		assert.True(t, photo.SetRating(4, SrcMeta))
		assert.Equal(t, 4, photo.PhotoRating)
		assert.Equal(t, SrcMeta, photo.RatingSrc)
	})

	t.Run("Clamp", func(t *testing.T) {
		photo := &Photo{}
		assert.True(t, photo.SetRating(9, SrcManual))
		assert.Equal(t, 5, photo.PhotoRating)
		assert.Equal(t, SrcManual, photo.RatingSrc)
	})

	t.Run("Priority", func(t *testing.T) {
		photo := &Photo{PhotoRating: 4, RatingSrc: SrcManual}
		assert.False(t, photo.SetRating(2, SrcXmp))
		assert.Equal(t, 4, photo.PhotoRating)
		assert.Equal(t, SrcManual, photo.RatingSrc)

		assert.True(t, photo.SetRating(5, SrcAdmin))
		assert.Equal(t, 5, photo.PhotoRating)
		assert.Equal(t, SrcAdmin, photo.RatingSrc)
	})

	t.Run("ReindexSameSource", func(t *testing.T) {
		photo := &Photo{PhotoRating: 2, RatingSrc: SrcXmp}
		assert.True(t, photo.SetRating(3, SrcXmp))
		assert.Equal(t, 3, photo.PhotoRating)
		assert.Equal(t, SrcXmp, photo.RatingSrc)
	})

	t.Run("ManualClear", func(t *testing.T) {
		photo := &Photo{PhotoRating: 4, RatingSrc: SrcXmp}
		assert.True(t, photo.SetRating(0, SrcManual))
		assert.Equal(t, 0, photo.PhotoRating)
		assert.Equal(t, SrcManual, photo.RatingSrc)
		assert.False(t, photo.SetRating(4, SrcXmp))
		assert.Equal(t, 0, photo.PhotoRating)
	})
}
