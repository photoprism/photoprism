package report

// Bool returns t or f, depending on the value of b.
func Bool(value bool, yes, no string) string {
	if value {
		return yes
	}

	return no
}
