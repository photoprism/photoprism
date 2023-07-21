package meta

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGpsToLat(t *testing.T) {
	lat := GpsToDecimal("51 deg 15' 17.47\" N")
	exp := 51.254852

	assert.InEpsilon(t, lat, exp, 0.1)
}

func TestGpsToLng(t *testing.T) {
	lng := GpsToDecimal("7 deg 23' 22.09\" E")
	exp := 7.389470

	assert.InEpsilon(t, lng, exp, 0.1)
}

func TestGpsToLatLng(t *testing.T) {
	t.Run("valid string", func(t *testing.T) {
		lat, lng := GpsToLatLng("51 deg 15' 17.47\" N, 7 deg 23' 22.09\" E")
		expLat, expLng := 51.254852, 7.389470

		assert.InEpsilon(t, lat, expLat, 0.1)
		assert.InEpsilon(t, lng, expLng, 0.1)
	})

	t.Run("empty string", func(t *testing.T) {
		lat, lng := GpsToLatLng("")
		assert.Equal(t, float64(0), lat)
		assert.Equal(t, float64(0), lng)
	})

	t.Run("invalid string", func(t *testing.T) {
		lat, lng := GpsToLatLng("abc bdf")
		assert.Equal(t, float64(0), lat)
		assert.Equal(t, float64(0), lng)
	})
}

func TestGpsToDecimal(t *testing.T) {
	t.Run("valid string", func(t *testing.T) {
		r := GpsToDecimal("51 deg 15' 17.47\" N")
		assert.InEpsilon(t, 51.25485277777778, r, 0.01)
	})

	t.Run("empty string", func(t *testing.T) {
		r := GpsToDecimal("")
		assert.Equal(t, float64(0), r)
	})

	t.Run("invalid string", func(t *testing.T) {
		r := GpsToDecimal("abc")
		assert.Equal(t, float64(0), r)
	})
}

func TestGpsCoord(t *testing.T) {
	t.Run("valid string", func(t *testing.T) {
		r := ParseFloat("51")
		assert.Equal(t, float64(51), r)
	})

	t.Run("empty string", func(t *testing.T) {
		r := ParseFloat("")
		assert.Equal(t, float64(0), r)
	})

	t.Run("invalid string", func(t *testing.T) {
		r := ParseFloat("abc")
		assert.Equal(t, float64(0), r)
	})
}

func TestClipLat(t *testing.T) {
	assert.Equal(t, 10.254852777777785, clipLat(100.25485277777778))
	assert.Equal(t, 89.25485277777778, clipLat(89.25485277777778))
	assert.Equal(t, 10.254852777777785, clipLat(190.25485277777778))
	assert.Equal(t, -10.254852777777785, clipLat(-100.25485277777778))
	assert.Equal(t, -89.25485277777778, clipLat(-89.25485277777778))
	assert.Equal(t, -10.254852777777785, clipLat(-190.25485277777778))
}

func TestNormalizeGPS(t *testing.T) {
	assert.Equal(t, 100.25485277777778, normalizeCoord(100.25485277777778, 120.25485277777778))
	assert.Equal(t, 110.25485277777778, normalizeCoord(-130.25485277777778, 120.25485277777778))
	assert.Equal(t, -120.25485277777778, normalizeCoord(120.25485277777778, 120.25485277777778))
}
