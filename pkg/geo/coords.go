package geo

// Geographic coordinate bounds in decimal degrees.
const (
	LatitudeMin  = -90.0
	LatitudeMax  = 90.0
	LongitudeMin = -180.0
	LongitudeMax = 180.0
)

// CoordinateBoundaryTolerance accepts tiny IEEE-754 overflow at GPS bounds.
const CoordinateBoundaryTolerance = 0.0001

// ClampCoordinateBounds clamps coordinates to their hard latitude/longitude limits.
func ClampCoordinateBounds(lat, lng float64) (float64, float64, bool) {
	changed := false

	if lat > LatitudeMax {
		lat = LatitudeMax
		changed = true
	} else if lat < LatitudeMin {
		lat = LatitudeMin
		changed = true
	}

	if lng > LongitudeMax {
		lng = LongitudeMax
		changed = true
	} else if lng < LongitudeMin {
		lng = LongitudeMin
		changed = true
	}

	return lat, lng, changed
}

// NormalizeCoordinateBounds clips minor coordinate overshoot at hard latitude/longitude limits.
//
// Values beyond the configured tolerance are intentionally left unchanged so callers can still
// reject clearly invalid input and avoid expensive downstream processing.
func NormalizeCoordinateBounds(lat, lng float64) (float64, float64, bool) {
	changed := false

	if lat > LatitudeMax && lat <= LatitudeMax+CoordinateBoundaryTolerance {
		lat = LatitudeMax
		changed = true
	} else if lat < LatitudeMin && lat >= LatitudeMin-CoordinateBoundaryTolerance {
		lat = LatitudeMin
		changed = true
	}

	if lng > LongitudeMax && lng <= LongitudeMax+CoordinateBoundaryTolerance {
		lng = LongitudeMax
		changed = true
	} else if lng < LongitudeMin && lng >= LongitudeMin-CoordinateBoundaryTolerance {
		lng = LongitudeMin
		changed = true
	}

	return lat, lng, changed
}
