package config

import (
	"net/url"
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
)

// CdnUrl returns the optional content delivery network URI without trailing slash.
func (c *Config) CdnUrl(res string) string {
	if c.options.CdnUrl == "" || c.options.CdnUrl == c.options.SiteUrl {
		return res
	}

	return strings.TrimRight(c.options.CdnUrl, "/") + res
}

// UseCdn checks if a Content Deliver Network (CDN) is used to serve static content.
func (c *Config) UseCdn() bool {
	if c.options.CdnUrl == "" || c.options.CdnUrl == c.options.SiteUrl {
		return false
	}

	return true
}

// NoCdn checks if there is no Content Deliver Network (CDN) configured to serve static content.
func (c *Config) NoCdn() bool {
	return !c.UseCdn()
}

// CdnDomain returns the content delivery network domain name if specified.
func (c *Config) CdnDomain() string {
	if c.options.CdnUrl == "" || c.options.CdnUrl == c.options.SiteUrl {
		return ""
	} else if u, err := url.Parse(c.options.CdnUrl); err != nil {
		return ""
	} else {
		return u.Hostname()
	}
}

// CdnVideo checks if videos should be streamed using the configured CDN.
func (c *Config) CdnVideo() bool {
	if c.options.CdnUrl == "" || c.options.CdnUrl == c.options.SiteUrl {
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
