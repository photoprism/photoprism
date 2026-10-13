package clean

// StripControl replaces ASCII whitespace controls with a space and removes the other C0 controls and DEL.
// It works on bytes, as these never occur inside a multibyte UTF-8 sequence.
func StripControl(s string) string {
	for i := 0; i < len(s); i++ {
		if !unsafeDropRune(rune(s[i])) {
			continue
		}

		dst := make([]byte, 0, len(s))
		dst = append(dst, s[:i]...)

		for j := i; j < len(s); j++ {
			switch b := s[j]; {
			case b == '\t', b == '\n', b == '\v', b == '\f', b == '\r':
				dst = append(dst, ' ')
			case unsafeDropRune(rune(b)):
				continue
			default:
				dst = append(dst, b)
			}
		}

		return string(dst)
	}

	return s
}
