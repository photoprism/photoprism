package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/time/tz"
)

func TestPhoto_SetRating(t *testing.T) {
	t.Run("Unrated", func(t *testing.T) {
		m := Photo{}

		assert.False(t, m.HasRating())
		assert.Equal(t, 0, m.GetRating())
		assert.Equal(t, "", m.GetRatingSrc())
	})
	t.Run("Meta", func(t *testing.T) {
		m := Photo{}

		m.SetRating(4, SrcMeta)

		assert.True(t, m.HasRating())
		assert.Equal(t, 4, m.GetRating())
		assert.Equal(t, SrcMeta, m.GetRatingSrc())
	})
	t.Run("LowerPriorityIsIgnored", func(t *testing.T) {
		m := Photo{PhotoRating: 4, RatingSrc: SrcMeta}

		m.SetRating(2, SrcEstimate)

		assert.Equal(t, 4, m.GetRating())
		assert.Equal(t, SrcMeta, m.GetRatingSrc())
	})
	t.Run("SamePriorityWins", func(t *testing.T) {
		m := Photo{PhotoRating: 4, RatingSrc: SrcMeta}

		m.SetRating(2, SrcMeta)

		assert.Equal(t, 2, m.GetRating())
	})
	t.Run("ExplicitZeroStaysDistinguishable", func(t *testing.T) {
		m := Photo{PhotoRating: 4, RatingSrc: SrcMeta}

		m.SetRating(0, SrcManual)

		assert.True(t, m.HasRating())
		assert.Equal(t, 0, m.GetRating())
		assert.Equal(t, SrcManual, m.GetRatingSrc())

		// A lower-priority source must not override the explicit 0 stars.
		m.SetRating(5, SrcMeta)

		assert.Equal(t, 0, m.GetRating())
		assert.Equal(t, SrcManual, m.GetRatingSrc())
	})
	t.Run("ManualWinsOverMeta", func(t *testing.T) {
		m := Photo{PhotoRating: 4, RatingSrc: SrcMeta}

		m.SetRating(3, SrcManual)

		assert.Equal(t, 3, m.GetRating())
		assert.Equal(t, SrcManual, m.GetRatingSrc())
	})
	t.Run("OutOfRangeIsIgnored", func(t *testing.T) {
		m := Photo{PhotoRating: 3, RatingSrc: SrcManual}

		m.SetRating(6, SrcAdmin)
		m.SetRating(-1, SrcAdmin)

		assert.Equal(t, 3, m.GetRating())
		assert.Equal(t, SrcManual, m.GetRatingSrc())
	})
}

func TestPhoto_NormalizeRating(t *testing.T) {
	t.Run("ClampAboveMax", func(t *testing.T) {
		m := Photo{TimeZone: tz.Local, PhotoRating: 9}

		assert.True(t, m.NormalizeValues())
		assert.Equal(t, 5, m.PhotoRating)
	})
	t.Run("ClampBelowMin", func(t *testing.T) {
		m := Photo{TimeZone: tz.Local, PhotoRating: -3}

		assert.True(t, m.NormalizeValues())
		assert.Equal(t, 0, m.PhotoRating)
	})
	t.Run("InRangeUnchanged", func(t *testing.T) {
		m := Photo{TimeZone: tz.Local, PhotoRating: 5}

		assert.False(t, m.NormalizeValues())
		assert.Equal(t, 5, m.PhotoRating)
	})
}
