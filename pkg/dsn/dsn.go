/*
Package dsn provides helpers for parsing database data source names, masking
credentials, and sharing driver-specific defaults used throughout PhotoPrism.

Copyright (c) 2018 - 2026 PhotoPrism UG. All rights reserved.

	This program is free software: you can redistribute it and/or modify
	it under Version 3 of the GNU Affero General Public License (the "AGPL"):
	<https://docs.photoprism.app/license/agpl>

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	The AGPL is supplemented by our Trademark and Brand Guidelines,
	which describe how our Brand Assets may be used:
	<https://www.photoprism.app/trademark/>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>
*/
package dsn

import (
	"net"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// dsnPattern is a regular expression matching a go mysql database DSN string.
var dsnPattern = regexp.MustCompile(`^((?P<driver>.*):\/\/)?(?:(?P<user>.*?)(?::(?P<password>.*))?@)?` +
	`(?:(?P<net>(?i)tcp|socket|pipe|memory|unix)(?:\((?P<server>[^\)]*)\))?)?` +
	`\/(?P<name>.*?)` +
	`(?:\?(?P<params>[^\?]*))?$`)

// dsnPostgresPasswordPattern is a regular expression matching a password in a PostgreSQL-style database DSN string.
var dsnPostgresPasswordPattern = regexp.MustCompile(`(?i)(password\s*=\s*)("[^"]*"|'[^']*'|\S+)`)

// DSN represents parts of a data source name.
type DSN struct {
	DSN      string
	Driver   string
	User     string
	Password string //nolint:gosec // G117: DSN component intentionally stores password.
	Net      string
	Server   string
	Name     string
	Params   string
}

// String returns the original DSN string.
func (d *DSN) String() string {
	return d.DSN
}

// SQLiteFilename returns the filename including any provided path from the DSN,
// or empty string if not applicable.
func (d DSN) SQLiteFilename() string {
	if d.Driver == DriverSQLite3 {
		if d.Server == "" {
			return d.Name
		} else {
			return filepath.Join(d.Server, d.Name)
		}
	} else {
		// Return an empty string, as this wasn't an SQLite DSN
		return ""
	}
}

// MaskPassword hides the password portion of a DSN while leaving the rest untouched for logging/reporting.
func (d *DSN) MaskPassword() (s string) {
	if d.DSN == "" || d.Password == "" {
		return d.DSN
	}

	s = d.DSN

	// Mask password in regular DSN.
	needle := ":" + d.Password + "@"
	if strings.Contains(s, needle) {
		return strings.Replace(s, needle, ":***@", 1)
	}

	// Mask password in PostgreSQL-style DSN.
	if d.Driver == DriverPostgres || strings.Contains(s, "password=") {
		return dsnPostgresPasswordPattern.ReplaceAllStringFunc(s, func(segment string) string {
			matches := dsnPostgresPasswordPattern.FindStringSubmatch(segment)
			if len(matches) != 3 {
				return segment
			}

			prefix := matches[1]
			value := matches[2]
			unquoted := strings.Trim(value, `'"`)

			if unquoted != d.Password {
				return segment
			}

			switch {
			case strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`):
				return prefix + `"` + "***" + `"`
			case strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`):
				return prefix + `'` + "***" + `'`
			default:
				return prefix + "***"
			}
		})
	}

	// Return DSN with masked password.
	return s
}

// Host the database server host.
func (d *DSN) Host() string {
	if d.Driver == DriverSQLite3 {
		return ""
	}

	host, _ := d.splitHostPort()
	return host
}

// Port the database server port.
func (d *DSN) Port() int {
	if d.Driver == DriverSQLite3 {
		return 0
	}

	defaultPort := 0

	switch d.Driver {
	case DriverMySQL, DriverMariaDB:
		defaultPort = 3306
	case DriverPostgres:
		defaultPort = 5432
	}

	if d.Server == "" {
		return 0
	}

	_, portValue := d.splitHostPort()

	if portValue == "" {
		return defaultPort
	}

	port, err := strconv.Atoi(portValue)
	if err != nil || port < 1 || port > 65535 {
		return defaultPort
	}

	return port
}

// splitHostPort splits the DSN server field into host and port components.
func (d *DSN) splitHostPort() (host, port string) {
	server := strings.TrimSpace(d.Server)

	if server == "" {
		return "", ""
	}

	var err error

	host, port, err = net.SplitHostPort(server)

	if err != nil {
		return server, ""
	}

	return host, port
}

// parse parses a data source name string.
// These are the following dsn formats that need to be supported
// with most of the components are optional.
// uri formats:
//
// file:[username[:password]@]database_name[?param1=value1&...&paramN=valueN]
// file:[username[:password]@]/path/to/database_name[?param1=value1&...&paramN=valueN]
// mysql://[username[:password]@]host[]:port]/database_name[?param1=value1&...&paramN=valueN]
// postgres://[username[:password]@][host][:port][/database_name][?param1=value1&...&paramN=valueN]
// postgressql://[username[:password]@][host][:port][/database_name][?param1=value1&...&paramN=valueN]
//
// go lang formats:
// mysql/mariadb
// [username[:password]@][protocol[(address)]]/[dbname][?param1=value1&...&paramN=valueN]
// sqlite
// database_name[?param1=value1&...&paramN=valueN]
// /path/to/database_name[?param1=value1&...&paramN=valueN]
//
// postgres format:
// keyword1=value1 keyword2=value2 ...

func (d *DSN) parse() {
	// Parse all URI based DSN 1st as they report the Driver and have a known format
	// driver://username:password@host:port/database_name?param=value&param=value
	if uri, err := url.Parse(d.DSN); err == nil {
		switch strings.ToLower(uri.Scheme) {
		case "file", "sqlite": // file is the Scheme that SQLite uses.
			// sqlite is an "alias" that needs to be recognized for tests to pass,
			// but it will fail to find a file if a connection is attempted.
			d.Driver = DriverSQLite3
			if uri.Opaque != "" {
				// SQLite supports a non standard URI style, so this has to be handled
				// driver:filepath/filename?param=value&param=value
				filePath := strings.TrimPrefix(uri.Opaque, uri.Scheme)
				d.Name = filepath.Base(filePath)
				d.Server = strings.TrimSuffix(filePath, d.Name)
				d.Params = uri.RawQuery
			} else {
				d.Name = filepath.Base(uri.Path)
				d.Server = strings.TrimSuffix(uri.Path, d.Name)
				d.User = uri.User.Username()
				d.Password, _ = uri.User.Password()
				d.Params = uri.RawQuery
			}
			if strings.ToLower(uri.Scheme) == "sqlite" {
				// Correct the incorrect scheme here, so that d.DSN can be used to connect with is so desired.
				d.DSN = strings.Replace(d.DSN, uri.Scheme, "file", 1)
			}
			d.detectDriver()
			return
		default:
			if uri.Opaque == "" && uri.Scheme != "" {
				// This has been recognized as a valid URI based dsn, so keep it.
				d.Driver = uri.Scheme
				d.Name = strings.TrimPrefix(uri.Path, "/")
				d.Server = uri.Host
				d.User = uri.User.Username()
				d.Password, _ = uri.User.Password()
				d.Params = uri.RawQuery
				d.detectDriver()
				return
			}
			// To be here, testing has shown that the username has been recognized as the uri.Schema
			// and the rest of the string is in the uri.Opaque and uri.RawQuery
			// username:password@protocol(address)/dbname?param=value&param=value
		}
	}

	// Try and parse as a key value pair Postgres dsn
	if d.parsePostgres() {
		return
	}

	// Try and parse as a MySQL/MariaDB GO format dsn
	if matches := dsnPattern.FindStringSubmatch(d.DSN); len(matches) > 0 {
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
	} else if len(d.DSN) > 0 {
		filePath, params, _ := strings.Cut(strings.TrimPrefix(d.DSN, "file:"), "?")
		d.Name = filepath.Base(filePath)
		d.Server = strings.TrimSuffix(filePath, d.Name)
		d.Params = params
	}

	d.detectDriver()
	// Add the missing / at the start of the file name if required for SQLite.
	if d.Driver == DriverSQLite3 {
		if d.Net != "" || d.User != "" || d.Password != "" {
			// Someone has managed to setup a filepath (with a name that ends in .db) that looks like a Go MySQL dsn
			filePath := strings.TrimSuffix(d.DSN, "?"+d.Params)
			d.Name = filepath.Base(filePath)
			d.Server = strings.TrimSuffix(filePath, d.Name)
			d.User = ""
			d.Password = ""
			d.Net = ""
		} else if strings.HasPrefix(d.DSN, "/") && d.Server == "" {
			filePath := "/" + d.Name
			d.Name = filepath.Base(filePath)
			d.Server = strings.TrimSuffix(filePath, d.Name)
		}
	}
}

// parsePostgres extracts connection settings from PostgreSQL key/value style DSNs and
// returns true on success.
func (d *DSN) parsePostgres() bool {
	if !strings.Contains(d.DSN, "password=") || !strings.Contains(d.DSN, "user=") {
		return false
	}

	fields, ok := d.splitKeyValue(d.DSN)

	if !ok {
		return false
	}

	values := make(map[string]string, len(fields))
	order := make([]string, 0, len(fields))

	for _, field := range fields {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) != 2 {
			return false
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if key == "" {
			return false
		}

		// Trim optional surrounding quotes.
		val = strings.Trim(val, `"`)

		values[key] = val
		order = append(order, key)
	}

	name := values["dbname"]

	if name == "" {
		if alt := values["database"]; alt != "" {
			name = alt
		} else {
			return false
		}
	}

	d.Driver = DriverPostgres
	d.User = values["user"]
	d.Password = values["password"]
	d.Name = name

	host := values["host"]
	port := values["port"]

	switch {
	case host != "" && port != "":
		d.Server = host + ":" + port
	case host != "":
		d.Server = host
	case port != "":
		d.Server = ":" + port
	}

	// Remove canonical keys so remaining values become Params.
	delete(values, "user")
	delete(values, "password")
	delete(values, "dbname")
	delete(values, "database")
	delete(values, "host")
	delete(values, "port")

	params := make([]string, 0, len(values))

	for _, key := range order {
		if val, ok := values[key]; ok {
			if strings.Contains(val, " ") {
				val = `"` + val + `"`
			}
			params = append(params, key+"="+val)
		}
	}

	if len(params) > 0 {
		d.Params = strings.Join(params, " ")
	}

	return true
}

// splitKeyValue tokenizes PostgreSQL key/value DSNs, supporting quoted values with spaces.
func (d *DSN) splitKeyValue(input string) ([]string, bool) {
	runes := []rune(strings.TrimSpace(input))

	if len(runes) == 0 || !strings.Contains(input, "=") {
		return nil, false
	}

	var (
		tokens    []string
		current   strings.Builder
		inQuotes  bool
		quoteRune rune
	)

	flush := func() {
		if current.Len() == 0 {
			return
		}
		tokens = append(tokens, current.String())
		current.Reset()
	}

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		switch {
		case inQuotes && r == '\\':
			if i+1 < len(runes) {
				current.WriteRune(runes[i+1])
				i++
			}
		case r == '\'' || r == '"':
			if inQuotes {
				if r == quoteRune {
					inQuotes = false
				} else {
					current.WriteRune(r)
				}
			} else {
				inQuotes = true
				quoteRune = r
			}
		case unicode.IsSpace(r):
			if inQuotes {
				current.WriteRune(r)
			} else {
				flush()
			}
		default:
			current.WriteRune(r)
		}
	}

	if inQuotes {
		return nil, false
	}

	flush()

	if len(tokens) == 0 {
		return nil, false
	}

	for _, token := range tokens {
		if !strings.Contains(token, "=") {
			return nil, false
		}
	}

	return tokens, true
}

// detectDriver infers the driver name from DSN contents. Explicit driver names
// flow through ParseDriver so both helpers share an alias set; unknown but
// non-empty names are kept (lowercased) rather than misclassified by the
// DSN-string heuristics below.
func (d *DSN) detectDriver() {
	if d.Driver != "" {
		if normalized := ParseDriver(d.Driver); normalized != "" {
			d.Driver = normalized
			return
		}
		d.Driver = strings.ToLower(d.Driver)
		return
	}

	lower := strings.ToLower(d.DSN)

	if strings.Contains(lower, "postgres://") || strings.Contains(lower, "postgresql://") {
		d.Driver = DriverPostgres
		return
	}

	if d.Net == "tcp" || d.Net == "unix" || strings.Contains(lower, "@tcp(") || strings.Contains(lower, "@unix(") {
		d.Driver = DriverMySQL
		return
	}

	if strings.HasPrefix(lower, "file:") || strings.HasPrefix(lower, ":memory:") || strings.HasSuffix(lower, ".db") ||
		strings.HasSuffix(strings.ToLower(d.Name), ".db") ||
		strings.Contains(strings.ToLower(d.Params), "_busy_timeout=") || strings.Contains(strings.ToLower(d.Params), "_journal_mode=") {
		d.Driver = DriverSQLite3
		return
	}

	if strings.Contains(lower, " host=") && strings.Contains(lower, " dbname=") {
		d.Driver = DriverPostgres
		return
	}

	if d.Server != "" && (strings.Contains(d.Server, ":") || d.Net != "") && d.Driver == "" {
		d.Driver = DriverMySQL
		return
	}

	if d.DSN == "/" {
		// The MySQL driver accepts just a slash as a valid DSN.  Nothing else does.
		d.Driver = DriverMySQL
		return
	}
}
