package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarkerMap_Get(t *testing.T) {
	t.Run("GetExistingMarker", func(t *testing.T) {
		r := MarkerFixtures.Get("1000003-3")
		assert.Equal(t, "Center", r.MarkerName)
		assert.Equal(t, float32(0.5), r.Y)
		assert.IsType(t, Marker{}, r)
	})
	t.Run("GetNotExistingMarker", func(t *testing.T) {
		r := MarkerFixtures.Get("XXX")
		assert.Equal(t, *UnknownMarker, r)
		assert.IsType(t, Marker{}, r)
	})
}

func TestMarkerMap_Pointer(t *testing.T) {
	t.Run("GetExistingMarkerPointer", func(t *testing.T) {
		r := MarkerFixtures.Pointer("1000003-3")
		assert.Equal(t, "Center", r.MarkerName)
		assert.Equal(t, float32(0.5), r.Y)
		assert.IsType(t, &Marker{}, r)
	})
	t.Run("GetNotExistingMarkerPointer", func(t *testing.T) {
		r := MarkerFixtures.Pointer("XXX")
		assert.Equal(t, UnknownMarker, r)
		assert.IsType(t, &Marker{}, r)
	})
}
