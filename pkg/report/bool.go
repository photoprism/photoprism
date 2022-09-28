package report

const (
	Enabled  = "enabled"
	Disabled = "disabled"
	Yes      = "yes"
	No       = "no"
)

// Bool returns t or f, depending on the value of b.
func Bool(b bool, t, f string) string {
	if b {
		return t
	}

	return f
}
