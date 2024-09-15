package geo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeg(t *testing.T) {
	t.Run("Lat0", func(t *testing.T) {
		dLat, dLng := Deg(0.0, 11.1)
		t.Logf("dLat: %f, dLng: %f", dLat, dLng)
		assert.InEpsilon(t, 0.000100, dLat, 0.01)
		assert.InEpsilon(t, 0.000100, dLng, 0.01)
	})
	t.Run("Lat23", func(t *testing.T) {
		dLat, dLng := Deg(23.0, 11.1)
		t.Logf("dLat: %f, dLng: %f", dLat, dLng)
		assert.InEpsilon(t, 0.000100, dLat, 0.01)
		assert.InEpsilon(t, 0.000108, dLng, 0.01)
	})
	t.Run("Lat45", func(t *testing.T) {
		dLat, dLng := Deg(45.0, 11.1)
		t.Logf("dLat: %f, dLng: %f", dLat, dLng)
		assert.InEpsilon(t, 0.000100, dLat, 0.01)
		assert.InEpsilon(t, 0.000141, dLng, 0.01)
	})
	t.Run("Lat67", func(t *testing.T) {
		dLat, dLng := Deg(67.0, 11.1)
		t.Logf("dLat: %f, dLng: %f", dLat, dLng)
		assert.InEpsilon(t, 0.000100, dLat, 0.01)
		assert.InEpsilon(t, 0.000255, dLng, 0.01)
	})
	t.Run("ZeroMeters", func(t *testing.T) {
		dLat, dLng := Deg(50.0, 0.0)
		t.Logf("dLat: %f, dLng: %f", dLat, dLng)
		assert.Equal(t, 0.0, dLat)
		assert.Equal(t, 0.0, dLng)
	})
	t.Run("LatOutOfRange", func(t *testing.T) {
		dLat, dLng := Deg(123.0, 11.1)
		t.Logf("dLat: %f, dLng: %f", dLat, dLng)
		assert.InEpsilon(t, 0.000100, dLat, 0.01)
		assert.InEpsilon(t, 0.057195, dLng, 0.01)
		dLat, dLng = Deg(-600.0, 11.1)
		t.Logf("dLat: %f, dLng: %f", dLat, dLng)
		assert.InEpsilon(t, 0.000100, dLat, 0.01)
		assert.InEpsilon(t, 0.057195, dLng, 0.01)
		dLat, dLng = Deg(-600.0, 11.1)
		t.Logf("dLat: %f, dLng: %f", dLat, dLng)
		assert.InEpsilon(t, 0.000100, dLat, 0.01)
		assert.InEpsilon(t, 0.057195, dLng, 0.01)
	})
}

func TestDegKm(t *testing.T) {
	t.Run("Lat0", func(t *testing.T) {
		dLat, dLng := DegKm(0.0, 0.0111)
		t.Logf("dLat: %f, dLng: %f", dLat, dLng)
		assert.InEpsilon(t, 0.000100, dLat, 0.01)
		assert.InEpsilon(t, 0.000100, dLng, 0.01)
	})
	t.Run("Lat23", func(t *testing.T) {
		dLat, dLng := DegKm(23.0, 0.0111)
		t.Logf("dLat: %f, dLng: %f", dLat, dLng)
		assert.InEpsilon(t, 0.000100, dLat, 0.01)
		assert.InEpsilon(t, 0.000108, dLng, 0.01)
	})
	t.Run("Lat45", func(t *testing.T) {
		dLat, dLng := DegKm(45.0, 0.0111)
		t.Logf("dLat: %f, dLng: %f", dLat, dLng)
		assert.InEpsilon(t, 0.000100, dLat, 0.01)
		assert.InEpsilon(t, 0.000141, dLng, 0.01)
	})
	t.Run("Lat67", func(t *testing.T) {
		dLat, dLng := DegKm(67.0, 0.0111)
		t.Logf("dLat: %f, dLng: %f", dLat, dLng)
		assert.InEpsilon(t, 0.000100, dLat, 0.01)
		assert.InEpsilon(t, 0.000255, dLng, 0.01)
	})
}

func TestKm(t *testing.T) {
	t.Run("BerlinShanghai", func(t *testing.T) {
		berlin := Position{Name: "Berlin", Lat: 52.5243700, Lng: 13.4105300}
		shanghai := Position{Name: "Shanghai", Lat: 31.2222200, Lng: 121.4580600}

		result := Km(berlin, shanghai)

		assert.Equal(t, 8396, int(result))
	})
}
