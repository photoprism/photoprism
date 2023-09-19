package clean

import (
	"fmt"
	"strings"

	"github.com/photoprism/photoprism/pkg/txt"
)

// GPSBounds parses the GPS bounds (Lat N, Lng E, Lat S, Lng W) and returns the coordinates if any.
func GPSBounds(bounds string) (latNorth, lngEast, latSouth, lngWest float32, err error) {
	if len(bounds) < 7 {
		return 0, 0, 0, 0, fmt.Errorf("no coordinates found")
	}

	values := strings.SplitN(bounds, ",", 5)
	found := len(values)

	// Invalid number of values?
	if found != 4 {
		return 0, 0, 0, 0, fmt.Errorf("invalid number of coordinates")
	}

	// Parse floating point coordinates.
	latNorth, lngEast, latSouth, lngWest = txt.Float32(values[0]), txt.Float32(values[1]), txt.Float32(values[2]), txt.Float32(values[3])

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

	if latNorth > latSouth {
		latNorth, latSouth = latSouth, latNorth
	}

	// Longitudes (from -180 to 180 degrees).
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

	if lngEast > lngWest {
		lngEast, lngWest = lngWest, lngEast
	}

	return latNorth, lngEast, latSouth, lngWest, nil
}
