package clean

// Orientation returns the Exif orientation value within a valid range or 0 if it is invalid.
func Orientation(val int) int {
	// Ignore invalid values.
	if val < 1 || val > 8 {
		return 0
	}

	return val
}
