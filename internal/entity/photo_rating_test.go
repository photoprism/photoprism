package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhoto_SetRating(t *testing.T) {
	t.Run("UnknownAndZero", func(t *testing.T) {
		assert.False(t, (&Photo{PhotoRating: PhotoRatingUnknown}).HasRating())
		assert.True(t, (&Photo{PhotoRating: 0, RatingSrc: SrcMeta}).HasRating())
	})

	t.Run("Initial", func(t *testing.T) {
		photo := &Photo{PhotoRating: PhotoRatingUnknown}

		assert.True(t, photo.SetRating(4, SrcMeta))
		assert.Equal(t, int8(4), photo.PhotoRating)
		assert.Equal(t, SrcMeta, photo.RatingSrc)
		assert.True(t, photo.HasRating())
	})

	t.Run("PreservesHigherPriority", func(t *testing.T) {
		photo := &Photo{PhotoRating: 5, RatingSrc: SrcManual}

		assert.False(t, photo.SetRating(2, SrcMeta))
		assert.Equal(t, int8(5), photo.PhotoRating)
		assert.Equal(t, SrcManual, photo.RatingSrc)
	})

	t.Run("AllowsHigherPriority", func(t *testing.T) {
		photo := &Photo{PhotoRating: 2, RatingSrc: SrcMeta}

		assert.True(t, photo.SetRating(3, SrcXmp))
		assert.Equal(t, int8(3), photo.PhotoRating)
		assert.Equal(t, SrcXmp, photo.RatingSrc)
	})

	t.Run("RejectsInvalid", func(t *testing.T) {
		photo := &Photo{PhotoRating: PhotoRatingUnknown}

		assert.False(t, photo.SetRating(6, SrcMeta))
		assert.Equal(t, int8(PhotoRatingUnknown), photo.PhotoRating)
		assert.Equal(t, "", photo.RatingSrc)
	})
}
