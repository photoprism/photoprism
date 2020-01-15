/*
Package s2 encapsulates Google's S2 library.

Additional information can be found in our Developer Guide:

https://github.com/photoprism/photoprism/wiki

...and in the Google S2 documentation:

https://s2geometry.io/

*/
package s2

import (
	gs2 "github.com/golang/geo/s2"
)

// Default cell level, see https://s2geometry.io/resources/s2cell_statistics.html.
var DefaultLevel = 21

// Token returns the S2 cell token for coordinates using the default level.
func Token(lat, lng float64) string {
	return TokenLevel(lat, lng, DefaultLevel)
}

// Token returns the S2 cell token for coordinates.
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
	if token == "" || token == "-" {
		return 0.0, 0.0
	}

	c := gs2.CellIDFromToken(token)

	if !c.IsValid() {
		return 0.0, 0.0
	}

	l := c.LatLng()
	return l.Lat.Degrees(), l.Lng.Degrees()
}

// IsZero returns true if the coordinates are both empty.
func IsZero(lat, lng float64) bool {
	return lat == 0.0 && lng == 0.0
}

// Range returns a token range for finding nearby locations.
func Range(token string, levelUp int) (min, max string) {
	c := gs2.CellIDFromToken(token)

	if !c.IsValid() {
		return min, max
	}

	// See https://s2geometry.io/resources/s2cell_statistics.html
	lvl := c.Level()

	parent := c.Parent(lvl - levelUp)

	return parent.Prev().ChildBeginAtLevel(lvl).ToToken(), parent.Next().ChildBeginAtLevel(lvl).ToToken()
}
