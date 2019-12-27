package maps

import olc "github.com/google/open-location-code/go"

var OlcLength = 8

func OlcEncode(lat, lng float64) string {
	if lat < -90 || lat > 90 {
		log.Warnf("olc: latitude out of range (%f)", lat)
		return ""
	}

	if lng < -180 || lng > 180 {
		log.Warnf("olc: longitude out of range (%f)", lng)
		return ""
	}

	return olc.Encode(lat, lng, OlcLength)
}
