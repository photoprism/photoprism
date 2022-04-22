package clean

import (
	"net/url"
	"strings"
)

// Uri removes invalid character from an uri string.
func Uri(s string) string {
	if s == "" || reject(s, 512) || strings.Contains(s, "..") {
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
