package clean

// Text normalizes line breaks to "\n", keeps tabs, replaces vertical tabs and form feeds with a space,
// and removes the other C0 controls and DEL, for multi-line text such as captions and notes.
// It works on bytes, as these never occur inside a multibyte UTF-8 sequence.
func Text(s string) string {
	for i := 0; i < len(s); i++ {
		if b := s[i]; b == '\t' || b == '\n' || !unsafeDropRune(rune(b)) {
			continue
		}

		dst := make([]byte, 0, len(s))
		dst = append(dst, s[:i]...)

		for j := i; j < len(s); j++ {
			switch b := s[j]; {
			case b == '\r':
				dst = append(dst, '\n')

				if j+1 < len(s) && s[j+1] == '\n' {
					j++
				}
			case b == '\v', b == '\f':
				dst = append(dst, ' ')
			case b == '\t', b == '\n':
				dst = append(dst, b)
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
