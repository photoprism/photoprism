package clean

// Token returns the sanitized token string with a length of up to 4096 characters.
// Allowed: [0-9a-zA-Z-_:] only. Fast path: return original when already valid ASCII.
func Token(s string) string {
	if s == "" || reject(s, LengthLimit) {
		return ""
	}

	// Fast path: check if all bytes are allowed ASCII.
	for i := 0; i < len(s); i++ {
		b := s[i]
		if !((b >= '0' && b <= '9') || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '-' || b == '_' || b == ':') {
			// Slow path: filter into a new byte slice.
			dst := make([]byte, 0, len(s))
			for j := 0; j < len(s); j++ {
				c := s[j]
				if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '-' || c == '_' || c == ':' {
					dst = append(dst, c)
				}
			}
			return string(dst)
		}
	}
	return s
}

// UrlToken returns the sanitized URL token with a length of up to 42 characters.
func UrlToken(s string) string {
	if s == "" || len(s) > 64 {
		return ""
	}

	return Token(s)
}

// ShareToken returns the sanitized link share token with a length of up to 160 characters.
func ShareToken(s string) string {
	if s == "" || len(s) > 160 {
		return ""
	}

	return Token(s)
}
