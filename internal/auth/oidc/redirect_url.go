package oidc

import (
	"errors"
	"net/url"
	"path"

	"github.com/photoprism/photoprism/internal/config"
)

// RedirectURL builds the OIDC redirect callback from the provided site URL.
func RedirectURL(siteUrl string) (string, error) {
	if siteUrl == "" {
		return "", errors.New("site url required")
	}

	u, err := url.Parse(siteUrl)

	if err != nil {
		return "", err
	}

	u.Path = path.Join(u.Path, config.OidcRedirectUri)

	return u.String(), nil
}

// CookiePath returns the URL path the OIDC RP state and PKCE cookies are scoped to:
// the OIDC endpoint base under the site URL's path, which covers both the
// /oidc/login leg and the /oidc/redirect callback. Scoping the cookies explicitly
// stops them from depending on a shared-domain reverse proxy rewriting the zitadel
// default Path=/ for the cookies to survive to the callback.
func CookiePath(siteUrl string) string {
	u, err := url.Parse(siteUrl)
	if err != nil {
		return "/"
	}

	return path.Dir(path.Join("/", u.Path, config.OidcRedirectUri))
}
