package dsn

import (
	"net"
	"net/url"
	"strings"

	"github.com/photoprism/photoprism/pkg/txt"
)

// PostgreSQL returns the DSN as a PostgreSQL connection URI, so that drivers read user names, passwords
// and database names with spaces or special characters as given. A server path starting with a slash is
// passed as the host parameter, which connects through the Unix socket in that directory, with port 5432
// unless Params sets one. Params are trusted and may be a URL query or space-separated key=value pairs.
func (d *DSN) PostgreSQL() string {
	query := url.Values{}

	if values, err := url.ParseQuery(d.Params); err == nil && (strings.Contains(d.Params, "&") || !strings.Contains(d.Params, " ")) {
		query = values
	} else {
		for _, pair := range strings.Fields(d.Params) {
			if key, value, found := strings.Cut(pair, "="); found && key != "" {
				query.Add(key, value)
			}
		}
	}

	u := url.URL{
		Scheme:  DriverPostgres,
		User:    url.UserPassword(d.User, d.Password),
		Path:    "/" + d.Name,
		RawPath: "/" + url.PathEscape(d.Name),
	}

	if strings.HasPrefix(d.Server, "/") {
		query.Set("host", d.Server)

		// The socket file name includes the port, which drivers otherwise take from the environment.
		if query.Get("port") == "" {
			query.Set("port", "5432")
		}
	} else {
		u.Host = d.Server
	}

	// Encode spaces as %20, which libpq tools decode as well; a literal "+" is already encoded as %2B.
	u.RawQuery = strings.ReplaceAll(query.Encode(), "+", "%20")

	return u.String()
}

// isPostgresURI reports whether a DSN is a PostgreSQL connection URI.
func isPostgresURI(s string) bool {
	lower := strings.ToLower(s)
	return strings.HasPrefix(lower, DriverPostgres+"://") || strings.HasPrefix(lower, "postgresql://")
}

// parsePostgresURI extracts the connection settings from a PostgreSQL connection URI, with percent-encoded
// values decoded, and returns true on success. User and password parameters override the user information,
// as they do for the drivers. Host and port parameters set the server, a socket directory followed by the port
// if there is one, e.g. "/var/run/postgresql:5433".
func (d *DSN) parsePostgresURI() bool {
	if !isPostgresURI(d.DSN) {
		return false
	}

	u, err := url.Parse(d.DSN)

	if err != nil {
		return false
	}

	query := u.Query()

	d.Driver = DriverPostgres
	d.User = u.User.Username()
	d.Password, _ = u.User.Password()
	d.Name = strings.TrimPrefix(u.Path, "/")
	d.Server = u.Host

	if user := query.Get("user"); user != "" {
		d.User = user
	}

	if password := query.Get("password"); password != "" {
		d.Password = password
	}

	// Host and port parameters take precedence, as they do for the drivers.
	host, port := query.Get("host"), query.Get("port")

	if port == "" {
		port = u.Port()
	}

	switch {
	case strings.HasPrefix(host, "/") && port != "":
		d.Server = host + ":" + port
	case strings.HasPrefix(host, "/"):
		d.Server = host
	case host != "" && port != "":
		d.Server = net.JoinHostPort(host, port)
	case host != "":
		d.Server = host
	case query.Get("port") != "":
		d.Server = net.JoinHostPort(u.Hostname(), port)
	}

	query.Del("host")
	query.Del("port")

	query.Del("user")
	query.Del("password")
	d.Params = query.Encode()

	return true
}

// maskPostgresURI masks the password in the user information and the password parameters of a PostgreSQL
// connection URI as written, so that percent-encoded and malformed values are masked as well. Rather than
// leave part of a password readable, it masks up to the last "@" if the host part does not contain one.
func maskPostgresURI(s string) string {
	_, rest, _ := strings.Cut(s, "://")
	start := len(s) - len(rest)
	host := rest

	if i := strings.Index(rest, "/"); i >= 0 {
		host = rest[:i]
	}

	at := strings.LastIndex(host, "@")

	if at < 0 {
		at = strings.LastIndex(rest, "@")
	}

	if at >= 0 {
		if colon := strings.Index(rest[:at], ":"); colon >= 0 {
			s = s[:start+colon+1] + txt.Masked + s[start+at:]
		}
	}

	// Parameters are separated by "&" only, and their names may be encoded or surrounded by spaces.
	base, query, found := strings.Cut(s, "?")

	if !found {
		return s
	}

	params := strings.Split(query, "&")

	for i, param := range params {
		key, _, hasValue := strings.Cut(param, "=")

		if name, err := url.QueryUnescape(key); err != nil {
			name = key
		} else if name = strings.ToLower(strings.TrimSpace(name)); hasValue && (name == "password" || name == "sslpassword") {
			params[i] = key + "=" + txt.Masked
		}
	}

	return base + "?" + strings.Join(params, "&")
}

// maskKeyValuePassword masks the value of a password match in a key/value DSN, keeping its quotes.
func maskKeyValuePassword(segment string) string {
	m := dsnKeyValuePasswordPattern.FindStringSubmatch(segment)

	if len(m) != 4 {
		return segment
	}

	switch value := m[3]; {
	case strings.HasPrefix(value, `"`):
		return m[1] + m[2] + `"` + txt.Masked + `"`
	case strings.HasPrefix(value, `'`):
		return m[1] + m[2] + `'` + txt.Masked + `'`
	default:
		return m[1] + m[2] + txt.Masked
	}
}
