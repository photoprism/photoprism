package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPlaceMap_Get(t *testing.T) {
	t.Run("get existing place", func(t *testing.T) {
		r := PlaceFixtures.Get("teotihuacan")
		assert.Equal(t, "Teotihuacán", r.LocCity)
		assert.Equal(t, "Mexico", r.LocState)
		assert.IsType(t, Place{}, r)
	})
	t.Run("get not existing place", func(t *testing.T) {
		r := PlaceFixtures.Get("xxx")
		assert.Equal(t, "Unknown", r.LocCity)
		assert.Equal(t, "Unknown", r.LocState)
		assert.IsType(t, Place{}, r)
	})
}

func TestPlaceMap_Pointer(t *testing.T) {
	t.Run("get existing place pointer", func(t *testing.T) {
		r := PlaceFixtures.Pointer("teotihuacan")
		assert.Equal(t, "Teotihuacán", r.LocCity)
		assert.Equal(t, "Mexico", r.LocState)
		assert.IsType(t, &Place{}, r)
	})
	t.Run("get not existing place pointer", func(t *testing.T) {
		r := PlaceFixtures.Pointer("xxx")
		assert.Equal(t, "Unknown", r.LocCity)
		assert.Equal(t, "Unknown", r.LocState)
		assert.IsType(t, &Place{}, r)
	})
}
