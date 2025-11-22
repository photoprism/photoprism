package txt

import (
	"errors"
	"regexp"
)

// PositionRegexp matches latitude/longitude pairs.
var PositionRegexp = regexp.MustCompile(`^([-+]?([1-8]?\d(\.\d+)?|90(\.0+)?)),\s*([-+]?(180(\.0+)?|((1[0-7]\d)|([1-9]?\d))(\.\d+)?))$`)

// Position parses a position string into latitude and longitude, or returns and error if the position is invalid.
func Position(s string) (lat, lng float64, err error) {
	pos := PositionRegexp.FindStringSubmatch(s)

	if len(pos) == 12 {
		lat = Float64(pos[1])
		lng = Float64(pos[5])

		if lat != 0.0 && lng != 0.0 {
			return lat, lng, nil
		}
	}

	return lat, lng, errors.New("invalid position")
}
