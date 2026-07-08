package config

import (
	"net/url"
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/http/scheme"
)

// CdnBaseUrl returns the configured CDN URL normalized as a base URL, or "" if unset.
// Strips the scheme's default port so semantically equal Site and CDN URLs compare equal.
func (c *Config) CdnBaseUrl() string {
	return scheme.NormalizeBaseURL(c.options.CdnUrl)
}

// CdnUrl returns the optional content delivery network URI without trailing slash.
func (c *Config) CdnUrl(res string) string {
	cdnUrl := c.CdnBaseUrl()

	if cdnUrl == "" || cdnUrl == c.SiteUrl() {
		return res
	}

	return strings.TrimRight(cdnUrl, "/") + res
}

// UseCdn checks if a Content Deliver Network (CDN) is used to serve static content.
func (c *Config) UseCdn() bool {
	cdnUrl := c.CdnBaseUrl()

	return cdnUrl != "" && cdnUrl != c.SiteUrl()
}

// NoCdn checks if there is no Content Deliver Network (CDN) configured to serve static content.
func (c *Config) NoCdn() bool {
	return !c.UseCdn()
}

// CdnDomain returns the content delivery network domain name if specified.
func (c *Config) CdnDomain() string {
	cdnUrl := c.CdnBaseUrl()

	if cdnUrl == "" || cdnUrl == c.SiteUrl() {
		return ""
	} else if u, err := url.Parse(cdnUrl); err != nil {
		return ""
	} else {
		return u.Hostname()
	}
}

// CdnVideo checks if videos should be streamed using the configured CDN.
func (c *Config) CdnVideo() bool {
	cdnUrl := c.CdnBaseUrl()

	if cdnUrl == "" || cdnUrl == c.SiteUrl() {
		return false
	}

	return c.options.CdnVideo
}

// CORSOrigin returns the value for the Access-Control-Allow-Origin header, if any.
func (c *Config) CORSOrigin() string {
	return clean.Header(c.options.CORSOrigin)
}

// CORSHeaders returns the value for the Access-Control-Allow-Headers header, if any.
func (c *Config) CORSHeaders() string {
	return clean.Header(c.options.CORSHeaders)
}

// CORSMethods returns the value for the Access-Control-Allow-Methods header, if any.
func (c *Config) CORSMethods() string {
	return clean.Header(c.options.CORSMethods)
}
