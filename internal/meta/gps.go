package meta

import (
	"math"
	"regexp"
	"strconv"

	"github.com/dsoprea/go-exif/v3"

	"github.com/photoprism/photoprism/pkg/clean"
)

// Latitude/Longitude bounds used when clamping map coordinates.
const (
	LatMax = 90
	LngMax = 180
)

// Regular expressions used to extract GPS coordinate components from EXIF strings.
var (
	GpsCoordsRegexp = regexp.MustCompile(`[0-9\.]+`)
	GpsRefRegexp    = regexp.MustCompile(`[NSEW]+`)
	GpsFloatRegexp  = regexp.MustCompile(`[+\-]?(?:(?:0|[1-9]\d*)(?:\.\d*)?|\.\d+)`)
)

// GpsToLatLng returns the GPS latitude and longitude as float point number.
func GpsToLatLng(s string) (lat, lng float64) {
	// Empty?
	if s == "" {
		return 0, 0
	}

	// Floating point numbers?
	if fl := GpsFloatRegexp.FindAllString(s, -1); len(fl) == 2 {
		if lat, err := strconv.ParseFloat(fl[0], 64); err != nil {
			log.Infof("metadata: %s is not a valid gps position", clean.Log(fl[0]))
		} else if lng, err := strconv.ParseFloat(fl[1], 64); err == nil {
			return lat, lng
		}
	}

	// Parse string values.
	co := GpsCoordsRegexp.FindAllString(s, -1)
	re := GpsRefRegexp.FindAllString(s, -1)

	if len(co) != 6 || len(re) != 2 {
		return 0, 0
	}

	latDeg := exif.GpsDegrees{
		Orientation: re[0][0],
		Degrees:     ParseFloat(co[0]),
		Minutes:     ParseFloat(co[1]),
		Seconds:     ParseFloat(co[2]),
	}

	lngDeg := exif.GpsDegrees{
		Orientation: re[1][0],
		Degrees:     ParseFloat(co[3]),
		Minutes:     ParseFloat(co[4]),
		Seconds:     ParseFloat(co[5]),
	}

	return latDeg.Decimal(), lngDeg.Decimal()
}

// GpsToDecimal returns the GPS latitude or longitude as a decimal
// floating-point number. Accepted forms: pure decimal ("47.6754"),
// 3-component DMS ("51 deg 15' 17.47\" N"), and 2-component
// degrees+decimal-minutes ("52,30.4567N", as Adobe XMP commonly writes).
func GpsToDecimal(s string) float64 {
	// Empty?
	if s == "" {
		return 0
	}

	// Floating point number?
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}

	// Parse string value.
	co := GpsCoordsRegexp.FindAllString(s, -1)
	re := GpsRefRegexp.FindAllString(s, -1)

	if len(re) != 1 {
		return 0
	}

	switch len(co) {
	case 2:
		// Adobe XMP 2-component form: degrees, decimal-minutes, cardinal
		// direction. Seconds are folded into the minutes value already.
		deg := exif.GpsDegrees{
			Orientation: re[0][0],
			Degrees:     ParseFloat(co[0]),
			Minutes:     ParseFloat(co[1]),
			Seconds:     0,
		}
		return deg.Decimal()
	case 3:
		// ExifTool / EXIF 3-component DMS form: degrees, minutes,
		// seconds, cardinal direction.
		deg := exif.GpsDegrees{
			Orientation: re[0][0],
			Degrees:     ParseFloat(co[0]),
			Minutes:     ParseFloat(co[1]),
			Seconds:     ParseFloat(co[2]),
		}
		return deg.Decimal()
	default:
		return 0
	}
}

// ParseFloat returns a single GPS coordinate value as floating point number (degree, minute or second).
func ParseFloat(s string) float64 {
	// Empty?
	if s == "" {
		return 0
	}

	// Parse floating point number.
	if result, err := strconv.ParseFloat(s, 64); err != nil {
		log.Debugf("metadata: %s is not a valid gps position", clean.Log(s))
		return 0
	} else {
		return result
	}
}

// NormalizeGPS normalizes the longitude and latitude of the GPS position to a generally valid range.
// Coordinates that are not finite numbers yield the zero position, which represents an unknown
// location downstream.
func NormalizeGPS(lat, lng float64) (float64, float64) {
	if !isFinite(lat) || !isFinite(lng) {
		return 0, 0
	}

	if lat < -LatMax || lat > LatMax || lng < -LngMax || lng >= LngMax {
		// Clip the latitude. Normalize the longitude.
		lat, lng = clipLat(lat), normalizeLng(lng)
	}

	return lat, lng
}

// isFinite reports whether a coordinate is a finite number.
func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func clipLat(lat float64) float64 {
	if lat > LatMax*2 {
		return math.Mod(lat, LatMax)
	} else if lat > LatMax {
		return lat - LatMax
	}

	if lat < -LatMax*2 {
		return math.Mod(lat, LatMax)
	} else if lat < -LatMax {
		return lat + LatMax
	}

	return lat
}

func normalizeLng(value float64) float64 {
	return normalizeCoord(value, LngMax)
}

// normalizeCoord returns a coordinate within [-max, max).
// A single modulo keeps the result independent of magnitude, as adding 2*max stops converging
// once that step falls below the representable precision. The in-range shortcut is a fast path
// rather than a correctness requirement, since math.Mod is exact below 2*max.
func normalizeCoord(value, max float64) float64 {
	if value >= -max && value < max {
		return value
	} else if !isFinite(value) {
		return 0
	}

	value = math.Mod(value, 2*max)

	switch {
	case value < -max:
		value += 2 * max
	case value >= max:
		value -= 2 * max
	case value == 0:
		// math.Mod keeps the sign of the dividend, so a negative multiple of 2*max gives -0.
		return 0
	}

	return value
}
