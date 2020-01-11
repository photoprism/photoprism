package s2

import (
	gs2 "github.com/golang/geo/s2"
	"github.com/photoprism/photoprism/internal/event"
)

var log = event.Log
var Level = 21

func Token(lat, lng float64) string {
	return TokenLevel(lat, lng, Level)
}

func TokenLevel(lat, lng float64, level int) string {
	if lat == 0.0 && lng == 0.0 {
		log.Debugf("s2: no values for latitude and longitude")
		return ""
	}

	if lat < -90 || lat > 90 {
		log.Warnf("s2: latitude out of range (%f)", lat)
		return ""
	}

	if lng < -180 || lng > 180 {
		log.Warnf("s2: longitude out of range (%f)", lng)
		return ""
	}

	l := gs2.LatLngFromDegrees(lat, lng)
	return gs2.CellIDFromLatLng(l).Parent(level).ToToken()
}

func LatLng(token string) (lat, lng float64) {
	if token == "" || token == "-" {
		log.Warn("s2: empty token")
		return 0.0, 0.0
	}

	c := gs2.CellIDFromToken(token)
	l := c.LatLng()
	return l.Lat.Degrees(), l.Lng.Degrees()
}
