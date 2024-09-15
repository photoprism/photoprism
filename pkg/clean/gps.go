package clean

import (
	"fmt"
	"strings"

	"github.com/photoprism/photoprism/pkg/geo"
	"github.com/photoprism/photoprism/pkg/txt"
)

// GPSBoundsDefaultPadding specifies the default padding of the GPS coordinates in meters.
const GPSBoundsDefaultPadding = 5.0

// GPSBounds parses the GPS bounds (Lat N, Lng E, Lat S, Lng W)
// and returns the coordinates with default padding.
func GPSBounds(bounds string) (latN, lngE, latS, lngW float64, err error) {
	return GPSBoundsWithPadding(bounds, GPSBoundsDefaultPadding)
}

// GPSBoundsWithPadding parses the GPS bounds (Lat N, Lng E, Lat S, Lng W)
// and returns the coordinates with a custom padding in meters.
func GPSBoundsWithPadding(bounds string, padding float64) (latN, lngE, latS, lngW float64, err error) {
	// Bounds string not long enough?
	if len(bounds) < 7 {
		return 0, 0, 0, 0, fmt.Errorf("no coordinates found")
	}

	// Trim whitespace and invalid characters.
	bounds = strings.Trim(bounds, " |\\<>\n\r\t\"'#$%!^*()[]{}")

	// Split bounding box string into coordinate values.
	values := strings.SplitN(bounds, ",", 5)
	found := len(values)

	// Return error if number of coordinates is invalid.
	if found != 4 {
		return 0, 0, 0, 0, fmt.Errorf("invalid number of coordinates")
	}

	// Convert coordinate strings to floating point values.
	latNorth, lngEast, latSouth, lngWest := txt.Float(values[0]), txt.Float(values[1]), txt.Float(values[2]), txt.Float(values[3])

	// Latitudes have a valid range of +90 to -90 degrees.
	if latNorth > 90 {
		latNorth = 90
	} else if latNorth < -90 {
		latNorth = -90
	}

	if latSouth > 90 {
		latSouth = 90
	} else if latSouth < -90 {
		latSouth = -90
	}

	// Make sure latSouth is smaller than latNorth.
	if latSouth > latNorth {
		latNorth, latSouth = latSouth, latNorth
	}

	// Longitudes have a valid range of -180 to +180 degrees.
	if lngEast > 180 {
		lngEast = 180
	} else if lngEast < -180 {
		lngEast = -180
	}

	if lngWest > 180 {
		lngWest = 180
	} else if lngWest < -180 {
		lngWest = -180
	}

	// Make sure lngWest is smaller than lngEast.
	if lngWest > lngEast {
		lngEast, lngWest = lngWest, lngEast
	}

	// Calculate the latitude and longitude padding in degrees.
	dLat, dLng := geo.Deg((latNorth+latSouth)/2.0, padding)

	// Return the coordinates of the bounding box with padding applied.
	return latNorth + dLat, lngEast + dLng, latSouth - dLat, lngWest - dLng, nil
}

// GPSLatRange returns a range based on the specified latitude and distance in km, or an error otherwise.
func GPSLatRange(lat float64, km float64) (latN, latS float64, err error) {
	// Latitude (from +90 to -90 degrees).
	if lat == 0 || lat < -90 || lat > 90 {
		return 0, 0, fmt.Errorf("invalid latitude")
	}

	// Approximate range radius.
	r := km * 0.75

	// Approximate longitude range,
	// see https://en.wikipedia.org/wiki/Decimal_degrees
	dLat, _ := geo.DegKm(lat, r)

	latN = lat + dLat
	latS = lat - dLat

	if latN > 90 {
		latN = 90
	}

	if latS < -90 {
		latS = -90
	}

	return latN, latS, nil
}

// GPSLngRange returns a range based on the specified longitude and distance in km, or an error otherwise.
func GPSLngRange(lat, lng float64, km float64) (lngE, lngW float64, err error) {
	// Longitude (from -180 to +180 degrees).
	if lng == 0 || lng < -180 || lng > 180 {
		return 0, 0, fmt.Errorf("invalid longitude")
	}

	// Approximate range radius.
	r := km * 0.75

	// Approximate longitude range,
	// see https://en.wikipedia.org/wiki/Decimal_degrees
	_, dLng := geo.DegKm(lat, r)

	lngE = lng + dLng
	lngW = lng - dLng

	if lngE > 180 {
		lngE = 180
	}

	if lngW < -180 {
		lngW = -180
	}

	return lngE, lngW, nil
}
