package meta

import (
	"math"
	"regexp"
	"strconv"

	"github.com/dsoprea/go-exif/v3"

	"github.com/photoprism/photoprism/pkg/clean"
)

const (
	LatMax = 90
	LngMax = 180
)

var GpsCoordsRegexp = regexp.MustCompile("[0-9\\.]+")
var GpsRefRegexp = regexp.MustCompile("[NSEW]+")
var GpsFloatRegexp = regexp.MustCompile("[+\\-]?(?:(?:0|[1-9]\\d*)(?:\\.\\d*)?|\\.\\d+)")

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

// GpsToDecimal returns the GPS latitude or longitude as decimal float point number.
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

	if len(co) != 3 || len(re) != 1 {
		return 0
	}

	latDeg := exif.GpsDegrees{
		Orientation: re[0][0],
		Degrees:     ParseFloat(co[0]),
		Minutes:     ParseFloat(co[1]),
		Seconds:     ParseFloat(co[2]),
	}

	return latDeg.Decimal()
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
func NormalizeGPS(lat, lng float64) (float32, float32) {
	if lat < LatMax || lat > LatMax || lng < LngMax || lng > LngMax {
		// Clip the latitude. Normalise the longitude.
		lat, lng = clipLat(lat), normalizeLng(lng)
	}

	return float32(lat), float32(lng)
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

func normalizeCoord(value, max float64) float64 {
	for value < -max {
		value += 2 * max
	}
	for value >= max {
		value -= 2 * max
	}
	return value
}
