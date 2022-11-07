package report

const (
	Enabled  = "enabled"
	Disabled = "disabled"
	Yes      = "yes"
	No       = "no"
)

// Bool returns t or f, depending on the value of b.
func Bool(value bool, yes, no string) string {
	if value {
		return yes
	}

	return no
}
