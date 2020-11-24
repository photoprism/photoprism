package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDescriptionMap_Get(t *testing.T) {
	t.Run("get existing description", func(t *testing.T) {
		r := DetailsFixtures.Get("lake", 1000000)
		assert.Equal(t, uint(1000000), r.PhotoID)
		assert.Equal(t, "nature, frog", r.Keywords)
		assert.IsType(t, Details{}, r)
	})
	t.Run("get not existing description", func(t *testing.T) {
		r := DetailsFixtures.Get("fantasy description", 123)
		assert.Equal(t, uint(123), r.PhotoID)
		assert.Equal(t, "", r.Keywords)
		assert.IsType(t, Details{}, r)
	})
}

func TestDescriptionMap_Pointer(t *testing.T) {
	t.Run("get existing description pointer", func(t *testing.T) {
		r := DetailsFixtures.Pointer("lake", 1000000)
		assert.Equal(t, uint(1000000), r.PhotoID)
		assert.Equal(t, "nature, frog", r.Keywords)
		assert.IsType(t, &Details{}, r)
	})
	t.Run("get not existing description pointer", func(t *testing.T) {
		r := DetailsFixtures.Pointer("fantasy 2", 345)
		assert.Equal(t, uint(345), r.PhotoID)
		assert.Equal(t, "", r.Keywords)
		assert.IsType(t, &Details{}, r)
	})
}
