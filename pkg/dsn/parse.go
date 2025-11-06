package dsn

// Parse creates a new DSN struct containing the parsed data from the specified string.
func Parse(dsn string) DSN {
	d := DSN{DSN: dsn}
	d.parse()
	return d
}
