package geo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeCoordinateBounds(t *testing.T) {
	t.Run("NoChange", func(t *testing.T) {
		lat, lng, changed := NormalizeCoordinateBounds(48.56344833333333, 8.996878333333333)
		assert.False(t, changed)
		assert.Equal(t, 48.56344833333333, lat)
		assert.Equal(t, 8.996878333333333, lng)
	})
	t.Run("ClampNorthEast", func(t *testing.T) {
		lat, lng, changed := NormalizeCoordinateBounds(90.000003, 180.000002)
		assert.True(t, changed)
		assert.Equal(t, 90.0, lat)
		assert.Equal(t, 180.0, lng)
	})
	t.Run("ClampSouthWest", func(t *testing.T) {
		lat, lng, changed := NormalizeCoordinateBounds(-90.000003, -180.000002)
		assert.True(t, changed)
		assert.Equal(t, -90.0, lat)
		assert.Equal(t, -180.0, lng)
	})
	t.Run("KeepLargeOverflow", func(t *testing.T) {
		lat, lng, changed := NormalizeCoordinateBounds(90.5, -181)
		assert.False(t, changed)
		assert.Equal(t, 90.5, lat)
		assert.Equal(t, -181.0, lng)
	})
}

func TestClampCoordinateBounds(t *testing.T) {
	t.Run("NoChange", func(t *testing.T) {
		lat, lng, changed := ClampCoordinateBounds(48.56344833333333, 8.996878333333333)
		assert.False(t, changed)
		assert.Equal(t, 48.56344833333333, lat)
		assert.Equal(t, 8.996878333333333, lng)
	})
	t.Run("ClampNorthEast", func(t *testing.T) {
		lat, lng, changed := ClampCoordinateBounds(22542883, 540)
		assert.True(t, changed)
		assert.Equal(t, 90.0, lat)
		assert.Equal(t, 180.0, lng)
	})
	t.Run("ClampSouthWest", func(t *testing.T) {
		lat, lng, changed := ClampCoordinateBounds(-22542883, -540)
		assert.True(t, changed)
		assert.Equal(t, -90.0, lat)
		assert.Equal(t, -180.0, lng)
	})
}
