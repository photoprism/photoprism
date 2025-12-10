package batch

// intToSafeUint casts an int to uint, returning the provided fallback when the value is negative.
func intToSafeUint(v int, fallback uint) uint {
	if v < 0 {
		return fallback
	}
	return uint(v)
}
