package clean

import (
	"fmt"
	"math"
	"strings"

	"github.com/photoprism/photoprism/pkg/geo"
	"github.com/photoprism/photoprism/pkg/txt"
)

// gpsCeil converts a GPS coordinate to a rounded float32 for use in queries.
func gpsCeil(f float64) float32 {
	return float32((math.Ceil(f*10000) / 10000) + 0.0001)
}

// gpsFloor converts a GPS coordinate to a rounded float32 for use in queries.
func gpsFloor(f float64) float32 {
	return float32((math.Floor(f*10000) / 10000) - 0.0001)
}

// GPSBounds parses the GPS bounds (Lat N, Lng E, Lat S, Lng W) and returns the coordinates if any.
func GPSBounds(bounds string) (latN, lngE, latS, lngW float32, err error) {
	// Bounds string not long enough?
	if len(bounds) < 7 {
		return 0, 0, 0, 0, fmt.Errorf("no coordinates found")
	}

	// Trim whitespace and invalid characters.
	bounds = strings.Trim(bounds, " |\\<>\n\r\t\"'#$%!^*()[]{}")

	// Split string into values.
	values := strings.SplitN(bounds, ",", 5)
	found := len(values)

	// Invalid number of values?
	if found != 4 {
		return 0, 0, 0, 0, fmt.Errorf("invalid number of coordinates")
	}

	// Parse floating point coordinates.
	latNorth, lngEast, latSouth, lngWest := txt.Float(values[0]), txt.Float(values[1]), txt.Float(values[2]), txt.Float(values[3])

	// Latitudes (from +90 to -90 degrees).
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

	// latSouth must be smaller.
	if latSouth > latNorth {
		latNorth, latSouth = latSouth, latNorth
	}

	// Longitudes (from -180 to +180 degrees).
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

	// lngWest must be smaller.
	if lngWest > lngEast {
		lngEast, lngWest = lngWest, lngEast
	}

	// Return rounded coordinates.
	return gpsCeil(latNorth), gpsCeil(lngEast), gpsFloor(latSouth), gpsFloor(lngWest), nil
}

// GPSLatRange returns a range based on the specified latitude and distance in km, or an error otherwise.
func GPSLatRange(lat float64, km uint) (latN, latS float32, err error) {
	// Latitude (from +90 to -90 degrees).
	if lat == 0 || lat < -90 || lat > 90 {
		return 0, 0, fmt.Errorf("invalid latitude")
	}

	// Approximate range.
	latN = gpsCeil(lat + geo.KmToDeg(km))
	latS = gpsFloor(lat - geo.KmToDeg(km))

	if latN > 90 {
		latN = 90
	}

	if latS < -90 {
		latS = -90
	}

	return latN, latS, nil
}

// GPSLngRange returns a range based on the specified longitude and distance in km, or an error otherwise.
func GPSLngRange(lng float64, km uint) (lngE, lngW float32, err error) {
	// Longitude (from -180 to +180 degrees).
	if lng == 0 || lng < -180 || lng > 180 {
		return 0, 0, fmt.Errorf("invalid longitude")
	}

	// Approximate range.
	lngE = gpsCeil(lng + geo.KmToDeg(km))
	lngW = gpsFloor(lng - geo.KmToDeg(km))

	if lngE > 180 {
		lngE = 180
	}

	if lngW < -180 {
		lngW = -180
	}

	return lngE, lngW, nil
}
