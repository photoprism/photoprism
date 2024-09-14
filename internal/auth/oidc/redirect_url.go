package oidc

import (
	"errors"
	"net/url"
	"path"

	"github.com/photoprism/photoprism/internal/config"
)

// RedirectURL returns the redirect URL for authentication via OIDC based on the specified site URL.
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
