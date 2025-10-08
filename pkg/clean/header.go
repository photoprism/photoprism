package clean

// Header sanitizes a string for use in request or response headers.
// Keeps printable ASCII (32..126). Fast path avoids allocation if unchanged.
func Header(s string) string {
	if s == "" || len(s) > LengthLimit {
		return ""
	}

	// Fast path: check if all bytes are already header-safe ASCII.
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b < 32 || b >= 127 {
			// Slow path: filter into a new byte slice.
			dst := make([]byte, 0, len(s))
			for j := 0; j < len(s); j++ {
				c := s[j]
				if c > 31 && c < 127 {
					dst = append(dst, c)
				}
			}
			return string(dst)
		}
	}
	return s
}
