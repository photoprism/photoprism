package clean

import (
	"net/url"
	"strings"
)

// Uri removes invalid character from an uri string.
func Uri(s string) string {
	if s == "" || len(s) > LengthLimit {
		return ""
	} else if strings.Contains(s, "..") {
		return ""
	}

	// Trim whitespace.
	s = strings.TrimSpace(s)

	if uri, err := url.Parse(s); err != nil {
		return ""
	} else {
		return uri.String()
	}
}

// UriRedacted removes credentials from a URI string while preserving non-sensitive components.
func UriRedacted(s string) string {
	if s == "" || len(s) > LengthLimit {
		return ""
	}

	// Trim whitespace.
	s = strings.TrimSpace(s)

	if uri, err := url.Parse(s); err != nil {
		return ""
	} else {
		return uri.Redacted()
	}
}
