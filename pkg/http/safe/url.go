package safe

import (
	"errors"
	"net/url"
	"strings"
)

var (
	// ErrURLHostRequired is returned when a URL does not include a hostname.
	ErrURLHostRequired = errors.New("missing URL host")
)

// URL parses a raw URL and ensures it uses HTTP(S) with a hostname.
func URL(rawURL string) (*url.URL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(strings.TrimSpace(u.Scheme)) {
	case "http", "https":
		if strings.TrimSpace(u.Host) == "" {
			return nil, ErrURLHostRequired
		}
		return u, nil
	default:
		return nil, ErrSchemeNotAllowed
	}
}
