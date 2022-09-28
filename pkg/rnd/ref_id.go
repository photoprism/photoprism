package rnd

// RefID generates a new reference ID with optional 2-character prefix.
func RefID(id string) string {
	if n := len(id); n == 0 {
		return "ref" + Base36(9)
	} else if n > 4 {
		return id[:4] + Base36(8)
	} else {
		return id + Base36(12-n)
	}
}

// IsRefID checks if the string is a valid reference ID.
func IsRefID(s string) bool {
	if n := len(s); n < 10 || n > 14 {
		return false
	}

	return IsAlnum(s)
}

// InvalidRefID checks if the reference ID is empty or invalid.
func InvalidRefID(s string) bool {
	return !IsRefID(s)
}
