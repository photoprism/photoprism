package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDescriptionMap_Get(t *testing.T) {
	t.Run("get existing description", func(t *testing.T) {
		r := DescriptionFixtures.Get("lake", 1000000)
		assert.Equal(t, uint(0xf4240), r.PhotoID)
		assert.Equal(t, "photo description lake", r.PhotoDescription)
		assert.IsType(t, Description{}, r)
	})
	t.Run("get not existing description", func(t *testing.T) {
		r := DescriptionFixtures.Get("fantasy description", 123)
		assert.Equal(t, uint(123), r.PhotoID)
		assert.Equal(t, "", r.PhotoDescription)
		assert.IsType(t, Description{}, r)
	})
}

func TestDescriptionMap_Pointer(t *testing.T) {
	t.Run("get existing description pointer", func(t *testing.T) {
		r := DescriptionFixtures.Pointer("lake", 1000000)
		assert.Equal(t, uint(0xf4240), r.PhotoID)
		assert.Equal(t, "photo description lake", r.PhotoDescription)
		assert.IsType(t, &Description{}, r)
	})
	t.Run("get not existing description pointer", func(t *testing.T) {
		r := DescriptionFixtures.Pointer("fantasy 2", 345)
		assert.Equal(t, uint(345), r.PhotoID)
		assert.Equal(t, "", r.PhotoDescription)
		assert.IsType(t, &Description{}, r)
	})
}
