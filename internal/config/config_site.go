package config

import (
	"fmt"
	"net/url"
	"strings"
)

// BaseUri returns the site base URI for a given resource.
func (c *Config) BaseUri(res string) string {
	if c.SiteUrl() == "" {
		return res
	}

	u, err := url.Parse(c.SiteUrl())

	if err != nil {
		return res
	}

	return strings.TrimRight(u.EscapedPath(), "/") + res
}

// ApiUri returns the api URI.
func (c *Config) ApiUri() string {
	return c.BaseUri(ApiUri)
}

// ContentUri returns the content delivery URI based on the CdnUrl and the ApiUri.
func (c *Config) ContentUri() string {
	return c.CdnUrl(c.ApiUri())
}

// VideoUri returns the video streaming URI.
func (c *Config) VideoUri() string {
	if c.CdnVideo() {
		return c.ContentUri()
	}

	return c.ApiUri()
}

// StaticUri returns the static content URI.
func (c *Config) StaticUri() string {
	return c.CdnUrl(c.BaseUri(StaticUri))
}

// StaticAssetUri returns the resource URI of the static file asset.
func (c *Config) StaticAssetUri(res string) string {
	return c.StaticUri() + "/" + res
}

// SiteUrl returns the public server URL (default is "http://localhost:2342/").
func (c *Config) SiteUrl() string {
	if c.options.SiteUrl == "" {
		return "http://localhost:2342/"
	}

	return strings.TrimRight(c.options.SiteUrl, "/") + "/"
}

// SiteHttps checks if the site URL uses HTTPS.
func (c *Config) SiteHttps() bool {
	if c.options.SiteUrl == "" {
		return false
	}

	return strings.HasPrefix(c.options.SiteUrl, "https://")
}

// SiteDomain returns the public server domain.
func (c *Config) SiteDomain() string {
	if u, err := url.Parse(c.SiteUrl()); err != nil {
		return "localhost"
	} else {
		return u.Hostname()
	}
}

// SiteAuthor returns the site author / copyright.
func (c *Config) SiteAuthor() string {
	return c.options.SiteAuthor
}

// SiteTitle returns the main site title (default is application name).
func (c *Config) SiteTitle() string {
	if c.options.SiteTitle == "" {
		return c.Name()
	}

	return c.options.SiteTitle
}

// SiteCaption returns a short site caption.
func (c *Config) SiteCaption() string {
	return c.options.SiteCaption
}

// SiteDescription returns a long site description.
func (c *Config) SiteDescription() string {
	return c.options.SiteDescription
}

// SitePreview returns the site preview image URL for sharing.
func (c *Config) SitePreview() string {
	if c.options.SitePreview == "" {
		return fmt.Sprintf("https://i.photoprism.app/prism?cover=64&style=centered%%20dark&caption=none&title=%s", url.QueryEscape(c.AppName()))
	}

	if !strings.HasPrefix(c.options.SitePreview, "http") {
		return c.SiteUrl() + strings.TrimPrefix(c.options.SitePreview, "/")
	}

	return c.options.SitePreview
}

// LegalInfo returns the legal info text for the page footer.
func (c *Config) LegalInfo() string {
	if s := c.CliGlobalString("imprint"); s != "" {
		log.Warnf("config: option 'imprint' is deprecated, please use 'legal-info'")
		return s
	}

	return c.options.LegalInfo
}

// LegalUrl returns the legal info url.
func (c *Config) LegalUrl() string {
	if s := c.CliGlobalString("imprint-url"); s != "" {
		log.Warnf("config: option 'imprint-url' is deprecated, please use 'legal-url'")
		return s
	}

	return c.options.LegalUrl
}
