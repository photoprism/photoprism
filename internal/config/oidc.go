package config

import "net/url"

func (c *Config) OidcIssuerUrl() *url.URL {
	if c.Options().OidcIssuerUrl == "" {
		return nil
	}
	res, err := url.Parse(c.Options().OidcIssuerUrl)
	if err != nil {
		log.Debugf("error parsing oidc issuer url: %q", err)
		return nil
	}
	return res
}

func (c *Config) OidcClientId() string {
	return c.Options().OidcClientID
}

func (c *Config) OidcClientSecret() string {
	return c.Options().OidcClientSecret
}
