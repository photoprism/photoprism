package meta

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGpsToLat(t *testing.T) {
	lat := GpsToDecimal("51 deg 15' 17.47\" N")
	exp := float32(51.254852)

	if lat-exp > 0 {
		t.Fatalf("lat is %f, should be %f", lat, exp)
	}
}

func TestGpsToLng(t *testing.T) {
	lng := GpsToDecimal("7 deg 23' 22.09\" E")
	exp := float32(7.389470)

	if lng-exp > 0 {
		t.Fatalf("lng is %f, should be %f", lng, exp)
	}
}

func TestGpsToLatLng(t *testing.T) {
	t.Run("valid string", func(t *testing.T) {
		lat, lng := GpsToLatLng("51 deg 15' 17.47\" N, 7 deg 23' 22.09\" E")
		expLat, expLng := float32(51.254852), float32(7.389470)

		if lat-expLat > 0 {
			t.Fatalf("lat is %f, should be %f", lat, expLat)
		}

		if lng-expLng > 0 {
			t.Fatalf("lng is %f, should be %f", lng, expLng)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		lat, lng := GpsToLatLng("")
		assert.Equal(t, float32(0), lat)
		assert.Equal(t, float32(0), lng)
	})

	t.Run("invalid string", func(t *testing.T) {
		lat, lng := GpsToLatLng("abc bdf")
		assert.Equal(t, float32(0), lat)
		assert.Equal(t, float32(0), lng)
	})
}

func TestGpsToDecimal(t *testing.T) {
	t.Run("valid string", func(t *testing.T) {
		r := GpsToDecimal("51 deg 15' 17.47\" N")
		assert.Equal(t, float32(51.254852), r)
	})

	t.Run("empty string", func(t *testing.T) {
		r := GpsToDecimal("")
		assert.Equal(t, float32(0), r)
	})

	t.Run("invalid string", func(t *testing.T) {
		r := GpsToDecimal("abc")
		assert.Equal(t, float32(0), r)
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
