package meta

import (
	"regexp"
	"strconv"

	"github.com/dsoprea/go-exif/v3"
)

// var GpsCoordsRegexp = regexp.MustCompile("(-?\\d+(\\.\\d+)?),\\s*(-?\\d+(\\.\\d+)?)")
var GpsCoordsRegexp = regexp.MustCompile("[0-9\\.]+")
var GpsRefRegexp = regexp.MustCompile("[NSEW]+")

// GpsToLatLng returns the GPS latitude and longitude as float point number.
func GpsToLatLng(s string) (lat, lng float32) {
	if s == "" {
		return 0, 0
	}

	co := GpsCoordsRegexp.FindAllString(s, -1)
	re := GpsRefRegexp.FindAllString(s, -1)

	if len(co) != 6 || len(re) != 2 {
		return 0, 0
	}

	latDeg := exif.GpsDegrees{
		Orientation: re[0][0],
		Degrees:     GpsCoord(co[0]),
		Minutes:     GpsCoord(co[1]),
		Seconds:     GpsCoord(co[2]),
	}

	lngDeg := exif.GpsDegrees{
		Orientation: re[1][0],
		Degrees:     GpsCoord(co[3]),
		Minutes:     GpsCoord(co[4]),
		Seconds:     GpsCoord(co[5]),
	}

	return float32(latDeg.Decimal()), float32(lngDeg.Decimal())
}

// GpsToDecimal returns the GPS latitude or longitude as decimal float point number.
func GpsToDecimal(s string) float32 {
	if s == "" {
		return 0
	}

	co := GpsCoordsRegexp.FindAllString(s, -1)
	re := GpsRefRegexp.FindAllString(s, -1)

	if len(co) != 3 || len(re) != 1 {
		return 0
	}

	latDeg := exif.GpsDegrees{
		Orientation: re[0][0],
		Degrees:     GpsCoord(co[0]),
		Minutes:     GpsCoord(co[1]),
		Seconds:     GpsCoord(co[2]),
	}

	return float32(latDeg.Decimal())
}

// GpsToLng returns a single GPS coordinate value as floating point number (degree, minute or second).
func GpsCoord(s string) float64 {
	if s == "" {
		return 0
	}

	result, err := strconv.ParseFloat(s, 64)

	if err != nil {
		log.Debugf("metadata: invalid floating-point number '%s' (gps coord)", s)
		return 0
	}

	return result
}
