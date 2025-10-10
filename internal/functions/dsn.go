package functions

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode"
)

// SQL Databases.
const (
	MySQL      = "mysql"
	MariaDB    = "mariadb"
	Postgres   = "postgres"
	PostgreSQL = "postgresql"
	SQLite3    = "sqlite"
)

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

	if len(matches) > 0 {
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
	} else {
		// Assume we have a PostgreSQL key value pair connection string.
		lastQuote := rune(0)
		smartSplit := func(char rune) bool {
			switch {
			case char == lastQuote:
				lastQuote = rune(0)
				return false
			case lastQuote != rune(0):
				return false
			case unicode.In(char, unicode.Quotation_Mark):
				lastQuote = char
				return false
			default:
				return unicode.IsSpace(char)
			}
		}
		pairs := strings.FieldsFunc(dsn, smartSplit)
		params := url.Values{}
		host := ""
		port := ""

		for _, pair := range pairs {
			splitPair := strings.Split(pair, "=")
			switch strings.ToLower(splitPair[0]) {
			case "host":
				host = splitPair[1]
			case "port":
				port = splitPair[1]
			case "user":
				d.User = splitPair[1]
			case "password":
				d.Password = splitPair[1]
			case "dbname":
				d.Name = splitPair[1]
			default:
				params.Add(splitPair[0], splitPair[1])
			}
		}
		d.Params = params.Encode()

		if len(host) > 0 && len(port) > 0 {
			d.Server = host + ":" + port
		} else if len(host) > 0 {
			d.Server = host
		} else {
			d.Server = ""
		}

		if len(pairs) > 1 {
			d.Driver = "postgresql"
		}
	}
}

// ToString returns the DSN in the format that gorm expects
func (d *DSN) ToString() string {
	driver := d.Driver
	if driver == "" {
		if d.User == "" {
			driver = SQLite3
		} else {
			driver = MariaDB
		}
	}

	switch driver {
	case SQLite3, "sqlitefile":
		if d.Params != "" {
			return fmt.Sprintf("%s/%s?%s", d.Server, d.Name, d.Params)
		} else {
			return fmt.Sprintf("%s/%s", d.Server, d.Name)
		}
	case PostgreSQL, Postgres:
		if d.Params != "" {
			return fmt.Sprintf("%s://%s:%s@%s/%s?%s", PostgreSQL, d.User, d.Password, d.Server, d.Name, d.Params)
		} else {
			return fmt.Sprintf("%s://%s:%s@%s/%s", PostgreSQL, d.User, d.Password, d.Server, d.Name)
		}
	case MariaDB, MySQL:
		databaseServer := d.Server
		if d.Net != "" {
			databaseServer = fmt.Sprintf("%s(%s)", d.Net, databaseServer)
		}
		if d.Params != "" {
			return fmt.Sprintf("%s:%s@%s/%s?%s", d.User, d.Password, databaseServer, d.Name, d.Params)
		} else {
			return fmt.Sprintf("%s:%s@%s/%s", d.User, d.Password, databaseServer, d.Name)
		}
	default:
		return ""
	}
}
