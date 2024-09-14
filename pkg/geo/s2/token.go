package s2

import (
	gs2 "github.com/golang/geo/s2"
)

// IsZero returns true if the coordinates are both empty.
func IsZero(lat, lng float64) bool {
	return lat == 0.0 && lng == 0.0
}

// Token returns the S2 cell token for coordinates using the default level.
func Token(lat, lng float64) string {
	return TokenLevel(lat, lng, DefaultLevel)
}

// TokenLevel returns the S2 cell token for coordinates.
func TokenLevel(lat, lng float64, level int) string {
	if lat == 0.0 && lng == 0.0 {
		return ""
	}

	if lat < -90 || lat > 90 {
		return ""
	}

	if lng < -180 || lng > 180 {
		return ""
	}

	l := gs2.LatLngFromDegrees(lat, lng)
	return gs2.CellIDFromLatLng(l).Parent(level).ToToken()
}

// LatLng returns the coordinates for a S2 cell token.
func LatLng(token string) (lat, lng float64) {
	token = NormalizeToken(token)

	if len(token) < 3 {
		return 0.0, 0.0
	}

	cell := gs2.CellIDFromToken(token)

	if !cell.IsValid() {
		return 0.0, 0.0
	}

	l := cell.LatLng()

	return l.Lat.Degrees(), l.Lng.Degrees()
}
