package clean

// ASCII removes all non-ASCII bytes from a string.
// Fast path: return the original string when it already contains only ASCII.
func ASCII(s string) string {
	if s == "" {
		return ""
	}

	// Fast path: all bytes < 128 → no allocation.
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 { // non-ASCII
			// Slow path: filter into a new byte slice.
			dst := make([]byte, 0, len(s))
			for j := 0; j < len(s); j++ {
				b := s[j]
				if b < 0x80 {
					dst = append(dst, b)
				}
			}
			return string(dst)
		}
	}
	return s
}
