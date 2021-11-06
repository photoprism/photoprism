package config

import "net/url"

func (c *Config) OidcIssuerUrl() *url.URL {
	if c.Options().OidcIssuerUrl == "" {
		return new(url.URL)
	}
	res, err := url.Parse(c.Options().OidcIssuerUrl)
	if err != nil {
		log.Debugf("error parsing oidc issuer url: %q", err)
		return new(url.URL)
	}
	return res
}

func (c *Config) OidcClientId() string {
	return c.Options().OidcClientID
}

func (c *Config) OidcClientSecret() string {
	return c.Options().OidcClientSecret
}

func (c *Config) OidcScopes() string {
	return c.Options().OidcScopes
}
