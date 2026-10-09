package dsn

import (
	"net/url"
	"strings"
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
