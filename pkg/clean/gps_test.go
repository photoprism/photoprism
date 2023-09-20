package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGPSBounds(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		latNorth, lngEast, latSouth, lngWest, err := GPSBounds("41.87760543823242,-87.62521362304688,41.89404296875,-87.6215591430664")
		assert.Equal(t, float32(41.8942), latNorth)
		assert.Equal(t, float32(41.8775), latSouth)
		assert.Equal(t, float32(-87.6254), lngWest)
		assert.Equal(t, float32(-87.6214), lngEast)
		assert.NoError(t, err)
	})
	t.Run("FlippedLat", func(t *testing.T) {
		latNorth, lngEast, latSouth, lngWest, err := GPSBounds("41.89404296875,-87.62521362304688,41.87760543823242,-87.6215591430664")
		assert.Equal(t, float32(41.8942), latNorth)
		assert.Equal(t, float32(41.8775), latSouth)
		assert.Equal(t, float32(-87.6254), lngWest)
		assert.Equal(t, float32(-87.6214), lngEast)
		assert.NoError(t, err)
	})
	t.Run("FlippedLng", func(t *testing.T) {
		latNorth, lngEast, latSouth, lngWest, err := GPSBounds("41.87760543823242,-87.6215591430664,41.89404296875,-87.62521362304688")
		assert.Equal(t, float32(41.8942), latNorth)
		assert.Equal(t, float32(41.8775), latSouth)
		assert.Equal(t, float32(-87.6254), lngWest)
		assert.Equal(t, float32(-87.6214), lngEast)
		assert.NoError(t, err)
	})
	t.Run("Empty", func(t *testing.T) {
		latNorth, lngEast, latSouth, lngWest, err := GPSBounds("")
		assert.Equal(t, float32(0), latNorth)
		assert.Equal(t, float32(0), lngEast)
		assert.Equal(t, float32(0), latSouth)
		assert.Equal(t, float32(0), lngWest)
		assert.Error(t, err)
	})
	t.Run("One", func(t *testing.T) {
		latNorth, lngEast, latSouth, lngWest, err := GPSBounds("41.87760543823242")
		assert.Equal(t, float32(0), latNorth)
		assert.Equal(t, float32(0), lngEast)
		assert.Equal(t, float32(0), latSouth)
		assert.Equal(t, float32(0), lngWest)
		assert.Error(t, err)
	})
	t.Run("Three", func(t *testing.T) {
		latNorth, lngEast, latSouth, lngWest, err := GPSBounds("41.87760543823242,-87.62521362304688,41.89404296875")
		assert.Equal(t, float32(0), latNorth)
		assert.Equal(t, float32(0), lngEast)
		assert.Equal(t, float32(0), latSouth)
		assert.Equal(t, float32(0), lngWest)
		assert.Error(t, err)
	})
	t.Run("Five", func(t *testing.T) {
		latNorth, lngEast, latSouth, lngWest, err := GPSBounds("41.87760543823242,-87.62521362304688,41.89404296875,-87.6215591430664,41.89404296875")
		assert.Equal(t, float32(0), latNorth)
		assert.Equal(t, float32(0), lngEast)
		assert.Equal(t, float32(0), latSouth)
		assert.Equal(t, float32(0), lngWest)
		assert.Error(t, err)
	})
}

func TestGPSLatRange(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		latNorth, latSouth, err := GPSLatRange(41.87760543823242, 2)
		assert.Equal(t, float32(41.8958), latNorth)
		assert.Equal(t, float32(41.8594), latSouth)
		assert.NoError(t, err)
	})
	t.Run("Zero", func(t *testing.T) {
		latNorth, latSouth, err := GPSLatRange(0, 2)
		assert.Equal(t, float32(0), latNorth)
		assert.Equal(t, float32(0), latSouth)
		assert.Error(t, err)
	})
}

func TestGPSLngRange(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		lngEast, lngWest, err := GPSLngRange(-87.62521362304688, 2)
		assert.Equal(t, float32(-87.6434), lngWest)
		assert.Equal(t, float32(-87.607), lngEast)
		assert.NoError(t, err)
	})
	t.Run("Zero", func(t *testing.T) {
		lngEast, lngWest, err := GPSLngRange(0, 2)
		assert.Equal(t, float32(0), lngEast)
		assert.Equal(t, float32(0), lngWest)
		assert.Error(t, err)
	})
}
