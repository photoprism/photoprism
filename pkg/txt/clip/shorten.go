package clip

// Shorten limits a character string to the specified number of runes and adds a suffix if it has been shortened.
func Shorten(s string, size int, suffix string) string {
	if suffix == "" {
		suffix = Ellipsis
	}

	l := len(suffix)

	if len(s) < size || size < l+1 {
		return s
	}

	return Runes(s, size-l) + suffix
}
