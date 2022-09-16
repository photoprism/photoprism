package config

import "regexp"

// dsnPattern is a regular expression matching a database DSN string.
var dsnPattern = regexp.MustCompile(
	`^((?P<driver>.*):\/\/)?(?:(?P<user>.*?)(?::(?P<password>.*))?@)?` +
		`(?:(?P<net>[^\(]*)(?:\((?P<server>[^\)]*)\))?)?` +
		`\/(?P<name>.*?)` +
		`(?:\?(?P<params>[^\?]*))?$`)

// DSN represents parts of a data source name.
type DSN struct {
	Driver   string
	User     string
	Password string
	Net      string
	Server   string
	Name     string
	Params   string
}

// NewDSN creates a new DSN struct from a string.
func NewDSN(dsn string) DSN {
	d := DSN{}
	d.Parse(dsn)
	return d
}

// Parse parses a data source name string.
func (d *DSN) Parse(dsn string) {
	if dsn == "" {
		return
	}

	matches := dsnPattern.FindStringSubmatch(dsn)
	names := dsnPattern.SubexpNames()

	for i, match := range matches {
		switch names[i] {
		case "driver":
			d.Driver = match
		case "user":
			d.User = match
		case "password":
			d.Password = match
		case "net":
			d.Net = match
		case "server":
			d.Server = match
		case "name":
			d.Name = match
		case "params":
			d.Params = match
		}
	}

	if d.Net != "" && d.Server == "" {
		d.Server = d.Net
		d.Net = ""
	}
}
