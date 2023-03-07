package clean

// Altitude returns the altitude within a maximum range as an integer, or 0 if it is invalid.
func Altitude(a float64) int {
	if a < -15000000 || a > 15000000 {
		return 0
	}

	return int(a)
}
