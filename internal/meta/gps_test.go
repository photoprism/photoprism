package meta

import (
	"fmt"
	"math"
	"testing"
	"time"

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

func TestIsFinite(t *testing.T) {
	t.Run("Number", func(t *testing.T) {
		assert.True(t, isFinite(0))
		assert.True(t, isFinite(-51.25))
		assert.True(t, isFinite(math.MaxFloat64))
	})
	t.Run("NaN", func(t *testing.T) {
		assert.False(t, isFinite(math.NaN()))
	})
	t.Run("Inf", func(t *testing.T) {
		assert.False(t, isFinite(math.Inf(1)))
		assert.False(t, isFinite(math.Inf(-1)))
	})
}

// TestNormalizeCoord covers the magnitudes at which a 2*max step falls below the representable
// precision, so a regression surfaces as a deadline rather than as a wrong value.
func TestNormalizeCoord(t *testing.T) {
	cases := []struct {
		name  string
		value float64
		want  float64
	}{
		{"InRange", 51.25, 51.25},
		{"LowerBound", -LngMax, -LngMax},
		{"UpperBound", LngMax, -LngMax},
		{"AboveRange", 190, -170},
		{"BelowRange", -190, 170},
		{"FullTurn", 360, 0},
		{"LargeFinite", 1e300, 0},
		{"BeyondStepSize", math.Pow(2, 63), 8},
		{"LargeFiniteWrapped", 1e17, -80},
		{"SlowConvergence", 1e15, -80},
		{"PosInf", math.Inf(1), 0},
		{"NegInf", math.Inf(-1), 0},
		{"NaN", math.NaN(), 0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, mustReturn(t, func() float64 { return normalizeCoord(c.value, LngMax) }))
		})
	}
}

// TestNormalizeGPSNonFinite covers that a position with a non-finite coordinate is reported as
// unknown rather than normalized to an arbitrary point.
func TestNormalizeGPSNonFinite(t *testing.T) {
	cases := []struct {
		name     string
		lat, lng float64
	}{
		{"InfLng", 48.5, math.Inf(1)},
		{"NegInfLng", 48.5, math.Inf(-1)},
		{"InfLat", math.Inf(1), 8.5},
		{"NaNLat", math.NaN(), 8.5},
		{"NaNLng", 48.5, math.NaN()},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			lat, lng := mustReturn2(t, func() (float64, float64) { return NormalizeGPS(c.lat, c.lng) })
			assert.Equal(t, float64(0), lat)
			assert.Equal(t, float64(0), lng)
		})
	}
	t.Run("ValidPosition", func(t *testing.T) {
		lat, lng := NormalizeGPS(48.5, 8.5)
		assert.Equal(t, 48.5, lat)
		assert.Equal(t, 8.5, lng)
	})
}

// TestNormalizeGPSRange covers the half-open longitude range: the upper bound wraps to the
// lower one, and a position inside the range is returned untouched.
func TestNormalizeGPSRange(t *testing.T) {
	t.Run("UpperBoundWraps", func(t *testing.T) {
		lat, lng := NormalizeGPS(LatMax, LngMax)
		assert.Equal(t, float64(LatMax), lat)
		assert.Equal(t, float64(-LngMax), lng)
	})
	t.Run("LowerBoundKept", func(t *testing.T) {
		lat, lng := NormalizeGPS(-LatMax, -LngMax)
		assert.Equal(t, float64(-LatMax), lat)
		assert.Equal(t, float64(-LngMax), lng)
	})
	t.Run("PoleWithInRangeLng", func(t *testing.T) {
		lat, lng := NormalizeGPS(LatMax, 100)
		assert.Equal(t, float64(LatMax), lat)
		assert.Equal(t, float64(100), lng)
	})
}

// TestNormalizeCoordSignOfZero covers that a full turn yields positive zero, since math.Mod
// would otherwise carry the sign of the dividend into the stored coordinate.
func TestNormalizeCoordSignOfZero(t *testing.T) {
	for _, v := range []float64{-2 * LngMax, -4 * LngMax, 2 * LngMax, 4 * LngMax} {
		t.Run(fmt.Sprintf("%v", v), func(t *testing.T) {
			assert.Equal(t, uint64(0), math.Float64bits(normalizeCoord(v, LngMax)))
		})
	}
	t.Run("NegativeZeroInput", func(t *testing.T) {
		v := math.Copysign(0, -1)
		assert.Equal(t, math.Float64bits(v), math.Float64bits(normalizeCoord(v, LngMax)))
	})
}

// mustReturn fails the test if fn does not return within a short deadline.
func mustReturn(t *testing.T, fn func() float64) float64 {
	t.Helper()
	done := make(chan float64, 1)
	go func() { done <- fn() }()

	select {
	case v := <-done:
		return v
	case <-time.After(5 * time.Second):
		t.Fatal("did not return")
		return 0
	}
}

// mustReturn2 fails the test if fn does not return within a short deadline.
func mustReturn2(t *testing.T, fn func() (float64, float64)) (float64, float64) {
	t.Helper()
	type pair struct{ a, b float64 }
	done := make(chan pair, 1)
	go func() { a, b := fn(); done <- pair{a, b} }()

	select {
	case v := <-done:
		return v.a, v.b
	case <-time.After(5 * time.Second):
		t.Fatal("did not return")
		return 0, 0
	}
}
