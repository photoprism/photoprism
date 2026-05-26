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
	t.Run("ValidString", func(t *testing.T) {
		lat, lng := GpsToLatLng("51 deg 15' 17.47\" N, 7 deg 23' 22.09\" E")
		expLat, expLng := 51.254852, 7.389470

		assert.InEpsilon(t, lat, expLat, 0.1)
		assert.InEpsilon(t, lng, expLng, 0.1)
	})
	t.Run("EmptyString", func(t *testing.T) {
		lat, lng := GpsToLatLng("")
		assert.Equal(t, float64(0), lat)
		assert.Equal(t, float64(0), lng)
	})
	t.Run("InvalidString", func(t *testing.T) {
		lat, lng := GpsToLatLng("abc bdf")
		assert.Equal(t, float64(0), lat)
		assert.Equal(t, float64(0), lng)
	})
}

func TestGpsToDecimal(t *testing.T) {
	t.Run("ValidString", func(t *testing.T) {
		r := GpsToDecimal("51 deg 15' 17.47\" N")
		assert.InEpsilon(t, 51.25485277777778, r, 0.01)
	})
	t.Run("EmptyString", func(t *testing.T) {
		r := GpsToDecimal("")
		assert.Equal(t, float64(0), r)
	})
	t.Run("InvalidString", func(t *testing.T) {
		r := GpsToDecimal("abc")
		assert.Equal(t, float64(0), r)
	})
	t.Run("PureDecimal", func(t *testing.T) {
		// Plain float passes through ParseFloat unchanged.
		assert.Equal(t, 47.6754, GpsToDecimal("47.6754"))
		assert.Equal(t, -47.6754, GpsToDecimal("-47.6754"))
	})
	t.Run("AdobeTwoComponentNorth", func(t *testing.T) {
		// 52° 30.4567'N → 52 + 30.4567/60 = 52.5076...
		r := GpsToDecimal("52,30.4567N")
		assert.InEpsilon(t, 52.50761166666667, r, 1e-6)
	})
	t.Run("AdobeTwoComponentSouth", func(t *testing.T) {
		// Cardinal S inverts the sign per exif.GpsDegrees.Decimal.
		r := GpsToDecimal("27,20.4263S")
		assert.InEpsilon(t, -27.340438333333333, r, 1e-6)
	})
	t.Run("AdobeTwoComponentEast", func(t *testing.T) {
		// 13° 24.5678'E → 13 + 24.5678/60 ≈ 13.4094633.
		r := GpsToDecimal("13,24.5678E")
		assert.InEpsilon(t, 13.409463333333334, r, 1e-6)
	})
	t.Run("AdobeTwoComponentLeadingZeros", func(t *testing.T) {
		// Adobe writes longitudes with leading zeros (031 = 31°).
		r := GpsToDecimal("031,53.5529E")
		assert.InEpsilon(t, 31.892548333333334, r, 1e-6)
	})
	t.Run("RejectsZeroComponentsWithRef", func(t *testing.T) {
		// One coordinate component plus a ref is too few to interpret.
		assert.Equal(t, float64(0), GpsToDecimal("N"))
	})
	t.Run("RejectsFourComponentsWithRef", func(t *testing.T) {
		// More than three components is also unsupported.
		assert.Equal(t, float64(0), GpsToDecimal("1 2 3 4 N"))
	})
}

// TestGpsToDecimal_RegressionAgainstExistingFixtures asserts that
// GpsToDecimal still parses the 3-component DMS form used by the
// JSON fixtures under testdata/, so a regression surfaces here before
// it reaches the broader exif/json test suites.
func TestGpsToDecimal_RegressionAgainstExistingFixtures(t *testing.T) {
	cases := []struct {
		name, input string
		want        float64
		eps         float64
	}{
		{"gopher-original Lat", `52 deg 27' 34.56" N`, 52.45960, 1e-4},
		{"gopher-original Lng", `13 deg 19' 18.48" E`, 13.32180, 1e-4},
		{"panorama360 Lat", `59 deg 50' 27.00" N`, 59.84083, 1e-4},
		{"panorama360 Lng", `30 deg 30' 36.00" E`, 30.51000, 1e-4},
		{"date.mov Lat", `55 deg 33' 48.96" N`, 55.56360, 1e-4},
		{"date.mov Lng", `37 deg 58' 56.64" E`, 37.98240, 1e-4},
		{"berlin-landscape Lat", `52 deg 27' 53.64" N`, 52.46490, 1e-4},
		{"berlin-landscape Lng", `13 deg 18' 53.28" E`, 13.31480, 1e-4},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.InDelta(t, c.want, GpsToDecimal(c.input), c.eps)
		})
	}
}

func TestGpsCoord(t *testing.T) {
	t.Run("ValidString", func(t *testing.T) {
		r := ParseFloat("51")
		assert.Equal(t, float64(51), r)
	})
	t.Run("EmptyString", func(t *testing.T) {
		r := ParseFloat("")
		assert.Equal(t, float64(0), r)
	})
	t.Run("InvalidString", func(t *testing.T) {
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
