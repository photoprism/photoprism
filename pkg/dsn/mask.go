package dsn

// Mask hides the password portion of a DSN while leaving the rest untouched for logging/reporting.
func Mask(dsn string) string {
	if dsn == "" {
		return ""
	}

	// Parse database DSN.
	d := Parse(dsn)

	// Return original DSN if no password was found.
	if d.Password == "" {
		return dsn
	}

	// Return DSN with masked password.
	return d.MaskPassword()
}
