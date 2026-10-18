package oauth

import (
	"errors"
	"net/url"
	"strings"

	"github.com/photoprism/photoprism/pkg/http/dns"
)

// ValidRedirectURI reports whether s may be used as a redirect URI: an HTTPS URL, an HTTP URL on a loopback
// host, or a private-use URI scheme in reverse domain notation, as used by native apps (RFC 8252).
func ValidRedirectURI(s string) bool {
	u, err := url.Parse(s)

	if err != nil || u.Scheme == "" || strings.Contains(s, "#") || u.User != nil || u.Opaque != "" {
		return false
	}

	switch strings.ToLower(u.Scheme) {
	case "https":
		return u.Hostname() != ""
	case "http":
		return dns.IsLoopbackHost(u.Hostname())
	default:
		return privateUseScheme(u.Scheme) && strings.HasPrefix(s[len(u.Scheme)+1:], "/") &&
			!strings.HasPrefix(s[len(u.Scheme)+1:], "//")
	}
}

// privateUseScheme reports whether scheme is in reverse domain notation, e.g. "app.photoprism.mobile".
func privateUseScheme(scheme string) bool {
	parts := strings.Split(scheme, ".")

	if len(parts) < 2 {
		return false
	}

	for _, part := range parts {
		if part == "" {
			return false
		}
	}

	return true
}

// MatchRedirectURI reports whether requested exactly matches one of the registered redirect URIs.
// For HTTP loopback redirect URIs, the port may differ, as native apps choose it at runtime (RFC 8252).
func MatchRedirectURI(registered []string, requested string) bool {
	if requested == "" || !ValidRedirectURI(requested) {
		return false
	}

	for _, uri := range registered {
		if uri != "" && (uri == requested || matchLoopbackURI(uri, requested)) {
			return true
		}
	}

	return false
}

// matchLoopbackURI reports whether two HTTP loopback URIs are equal except for the port.
func matchLoopbackURI(registered, requested string) bool {
	r, err := url.Parse(registered)

	if err != nil || r.Scheme != "http" || !dns.IsLoopbackHost(r.Hostname()) {
		return false
	}

	q, err := url.Parse(requested)

	if err != nil || q.Scheme != "http" || !dns.IsLoopbackHost(q.Hostname()) {
		return false
	}

	return r.Hostname() == q.Hostname() && r.EscapedPath() == q.EscapedPath() && r.RawQuery == q.RawQuery
}

// RedirectURL returns redirectURI with the specified response parameters added to its query string,
// keeping the well-formed query parameters it already has. A parameter with an empty value is removed.
func RedirectURL(redirectURI string, params url.Values) (string, error) {
	if !ValidRedirectURI(redirectURI) {
		return "", errors.New("oauth: invalid redirect uri")
	}

	u, err := url.Parse(redirectURI)

	if err != nil {
		return "", err
	}

	q := u.Query()

	for k, v := range params {
		if len(v) > 0 && v[0] != "" {
			q.Set(k, v[0])
		} else {
			q.Del(k)
		}
	}

	u.RawQuery = q.Encode()

	return u.String(), nil
}

// CodeRedirectURL returns the redirect URL that passes an authorization code to the client (RFC 9207).
func CodeRedirectURL(redirectURI, code, state, issuer string) (string, error) {
	if code == "" || issuer == "" {
		return "", errors.New("oauth: code and issuer required")
	}

	return RedirectURL(redirectURI, url.Values{"code": {code}, "state": {state}, "iss": {issuer}})
}

// ErrorRedirectURL returns the redirect URL that passes an error to the client (RFC 9207).
func ErrorRedirectURL(redirectURI, errCode, description, state, issuer string) (string, error) {
	if errCode == "" || issuer == "" {
		return "", errors.New("oauth: error code and issuer required")
	}

	return RedirectURL(redirectURI, url.Values{"error": {errCode}, "error_description": {description}, "state": {state}, "iss": {issuer}})
}
