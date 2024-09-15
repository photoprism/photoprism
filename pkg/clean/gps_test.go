package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGPSBoundsWithPadding(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		latNorth, lngEast, latSouth, lngWest, err := GPSBoundsWithPadding("41.87760543823242,-87.62521362304688,41.89404296875,-87.6215591430664", 1000)
		assert.InEpsilon(t, 41.903036, latNorth, 0.00001)
		assert.InEpsilon(t, -87.609479, lngEast, 0.00001)
		assert.InEpsilon(t, 41.868612, latSouth, 0.00001)
		assert.InEpsilon(t, -87.637294, lngWest, 0.00001)
		assert.NoError(t, err)
	})
}

func TestGPSBounds(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		latNorth, lngEast, latSouth, lngWest, err := GPSBounds("41.87760543823242,-87.62521362304688,41.89404296875,-87.6215591430664")
		assert.InEpsilon(t, 41.8942, latNorth, 0.00001)
		assert.InEpsilon(t, -87.6214, lngEast, 0.00001)
		assert.InEpsilon(t, 41.8775, latSouth, 0.00001)
		assert.InEpsilon(t, -87.6254, lngWest, 0.00001)
		assert.NoError(t, err)
	})
	t.Run("China", func(t *testing.T) {
		// Actual postion: Lat 39.8922, Lng 116.315
		latNorth, lngEast, latSouth, lngWest, err := GPSBounds("39.8922004699707,116.31500244140625,39.8922004699707,116.31500244140625")
		assert.InEpsilon(t, 39.8924, latNorth, 0.00001)
		assert.InEpsilon(t, 116.3152, lngEast, 0.00001)
		assert.InEpsilon(t, 39.8921, latSouth, 0.00001)
		assert.InEpsilon(t, 116.3149, lngWest, 0.00001)
		assert.NoError(t, err)
	})
	t.Run("FlippedLat", func(t *testing.T) {
		latNorth, lngEast, latSouth, lngWest, err := GPSBounds("41.89404296875,-87.62521362304688,41.87760543823242,-87.6215591430664")
		assert.InEpsilon(t, 41.8942, latNorth, 0.00001)
		assert.InEpsilon(t, -87.6214, lngEast, 0.00001)
		assert.InEpsilon(t, 41.8775, latSouth, 0.00001)
		assert.InEpsilon(t, -87.6254, lngWest, 0.00001)
		assert.NoError(t, err)
	})
	t.Run("FlippedLng", func(t *testing.T) {
		latNorth, lngEast, latSouth, lngWest, err := GPSBounds("41.87760543823242,-87.6215591430664,41.89404296875,-87.62521362304688")
		assert.InEpsilon(t, 41.8942, latNorth, 0.00001)
		assert.InEpsilon(t, -87.6214, lngEast, 0.00001)
		assert.InEpsilon(t, 41.8775, latSouth, 0.00001)
		assert.InEpsilon(t, -87.6254, lngWest, 0.00001)
		assert.NoError(t, err)
	})
	t.Run("Empty", func(t *testing.T) {
		latNorth, lngEast, latSouth, lngWest, err := GPSBounds("")
		assert.Equal(t, 0.0, latNorth)
		assert.Equal(t, 0.0, lngEast)
		assert.Equal(t, 0.0, latSouth)
		assert.Equal(t, 0.0, lngWest)
		assert.Error(t, err)
	})
	t.Run("One", func(t *testing.T) {
		latNorth, lngEast, latSouth, lngWest, err := GPSBounds("41.87760543823242")
		assert.Equal(t, 0.0, latNorth)
		assert.Equal(t, 0.0, lngEast)
		assert.Equal(t, 0.0, latSouth)
		assert.Equal(t, 0.0, lngWest)
		assert.Error(t, err)
	})
	t.Run("Three", func(t *testing.T) {
		latNorth, lngEast, latSouth, lngWest, err := GPSBounds("41.87760543823242,-87.62521362304688,41.89404296875")
		assert.Equal(t, 0.0, latNorth)
		assert.Equal(t, 0.0, lngEast)
		assert.Equal(t, 0.0, latSouth)
		assert.Equal(t, 0.0, lngWest)
		assert.Error(t, err)
	})
	t.Run("Five", func(t *testing.T) {
		latNorth, lngEast, latSouth, lngWest, err := GPSBounds("41.87760543823242,-87.62521362304688,41.89404296875,-87.6215591430664,41.89404296875")
		assert.Equal(t, 0.0, latNorth)
		assert.Equal(t, 0.0, lngEast)
		assert.Equal(t, 0.0, latSouth)
		assert.Equal(t, 0.0, lngWest)
		assert.Error(t, err)
	})
	t.Run("Invalid", func(t *testing.T) {
		latNorth, lngEast, latSouth, lngWest, err := GPSBounds("95.87760543823242,-197.62521362304688,98.89404296875,-197.6215591430664")
		assert.InEpsilon(t, 90.000045, latNorth, 0.00001)
		assert.InEpsilon(t, -179.974236, lngEast, 0.00001)
		assert.InEpsilon(t, 89.999955, latSouth, 0.00001)
		assert.InEpsilon(t, -180.025764, lngWest, 0.00001)
		assert.NoError(t, err)
	})
	t.Run("Invalid2", func(t *testing.T) {
		latNorth, lngEast, latSouth, lngWest, err := GPSBounds("-95.87760543823242,197.62521362304688,-98.89404296875,197.6215591430664")
		assert.InEpsilon(t, -89.999955, latNorth, 0.00001)
		assert.InEpsilon(t, 180.025764, lngEast, 0.00001)
		assert.InEpsilon(t, -90.000045, latSouth, 0.00001)
		assert.InEpsilon(t, 179.974236, lngWest, 0.00001)
		assert.NoError(t, err)
	})
}

func TestGPSLatRange(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		latNorth, latSouth, err := GPSLatRange(41.87760543823242, 2)
		assert.Equal(t, float32(41.891094), float32(latNorth))
		assert.Equal(t, float32(41.864117), float32(latSouth))
		assert.NoError(t, err)
	})
	t.Run("Zero", func(t *testing.T) {
		latNorth, latSouth, err := GPSLatRange(0, 2)
		assert.Equal(t, 0.0, latNorth)
		assert.Equal(t, 0.0, latSouth)
		assert.Error(t, err)
	})
}

func TestGPSLngRange(t *testing.T) {
	t.Run("Lat0", func(t *testing.T) {
		lngEast, lngWest, err := GPSLngRange(0.0, -87.62521362304688, 2)
		assert.Equal(t, float32(-87.6387), float32(lngWest))
		assert.Equal(t, float32(-87.611725), float32(lngEast))
		assert.NoError(t, err)
	})
	t.Run("Lat45", func(t *testing.T) {
		lngEast, lngWest, err := GPSLngRange(45.0, -87.62521362304688, 2)
		assert.Equal(t, float32(-87.644295), float32(lngWest))
		assert.Equal(t, float32(-87.60613), float32(lngEast))
		assert.NoError(t, err)
	})
	t.Run("Lat67", func(t *testing.T) {
		lngEast, lngWest, err := GPSLngRange(67.0, -87.62521362304688, 2)
		assert.Equal(t, float32(-87.65974), float32(lngWest))
		assert.Equal(t, float32(-87.59069), float32(lngEast))
		assert.NoError(t, err)
	})
	t.Run("Zero", func(t *testing.T) {
		lngEast, lngWest, err := GPSLngRange(0, 0, 2)
		assert.Equal(t, 0.0, lngEast)
		assert.Equal(t, 0.0, lngWest)
		assert.Error(t, err)
	})
}
