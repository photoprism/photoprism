package geo

import (
	"math"
)

const (
	DistLimit      float64 = 5000
	ScopeDistLimit float64 = 50
	DefaultDist    float64 = 2
)

// Deg returns the distance in decimal degrees based on the specified distance in meters and the latitude,
// see https://en.wikipedia.org/wiki/Decimal_degrees#Precision.
func Deg(lat, meter float64) (dLat, dLng float64) {
	if meter <= 0.0 {
		return 0, 0
	}

	// Calculate latitude distance in degrees.
	dLat = (meter / AverageEarthRadiusMeter) * (180.0 / math.Pi)

	// Do not calculate the exact longitude distance in
	// degrees if the latitude is zero or out of range.
	if lat == 0.0 {
		return dLat, dLat
	} else if lat < -89.9 {
		lat = -89.9
	} else if lat > 89.9 {
		lat = 89.9
	}

	// Calculate longitude distance in degrees.
	dLng = (meter / AverageEarthRadiusMeter) * (180.0 / math.Pi) / math.Cos(lat*math.Pi/180.0)

	return dLat, dLng
}

// DegKm returns the distance in decimal degrees based on the specified distance in kilometers and the latitude.
func DegKm(lat, km float64) (dLat, dLng float64) {
	return Deg(lat, km*1000.0)
}

// DegToRad converts a value from degrees to radians.
func DegToRad(d float64) float64 {
	return d * math.Pi / 180
}

// Km returns the shortest path between two positions in km.
func Km(p, q Position) (km float64) {
	if p.Lat == q.Lat && p.Lng == q.Lng {
		return 0.0
	}

	lat1 := DegToRad(p.Lat)
	lng1 := DegToRad(p.Lng)
	lat2 := DegToRad(q.Lat)
	lng2 := DegToRad(q.Lng)

	diffLat := lat2 - lat1
	diffLng := lng2 - lng1

	a := math.Pow(math.Sin(diffLat/2), 2) + math.Cos(lat1)*math.Cos(lat2)*
		math.Pow(math.Sin(diffLng/2), 2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return c * AverageEarthRadiusKm
}
