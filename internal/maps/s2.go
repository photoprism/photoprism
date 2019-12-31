package maps

import (
	"github.com/golang/geo/s2"
)

var S2Level = 15

func S2Token(lat, lng float64) string {
	return S2TokenLevel(lat, lng, S2Level)
}

func S2TokenLevel(lat, lng float64, level int) string {
	if lat < -90 || lat > 90 {
		log.Warnf("olc: latitude out of range (%f)", lat)
		return ""
	}

	if lng < -180 || lng > 180 {
		log.Warnf("olc: longitude out of range (%f)", lng)
		return ""
	}

	l := s2.LatLngFromDegrees(lat, lng)
	return s2.CellIDFromLatLng(l).Parent(level).ToToken()
}
