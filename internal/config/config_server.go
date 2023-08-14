package config

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/photoprism/photoprism/internal/server/header"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
)

const (
	HttpModeProd  = "release"
	HttpModeDebug = "debug"
)

// DetachServer checks if server should detach from console (daemon mode).
func (c *Config) DetachServer() bool {
	return c.options.DetachServer
}

// TrustedProxy returns the ranges from which reverse proxy headers can be trusted as comma-separated list.
func (c *Config) TrustedProxy() string {
	return strings.Join(c.options.TrustedProxies, ", ")
}

// TrustedProxies returns proxy server ranges from which reverse proxy headers can be trusted.
func (c *Config) TrustedProxies() []string {
	return c.options.TrustedProxies
}

// ProxyProtoHeader returns the proxy protocol header names.
func (c *Config) ProxyProtoHeader() []string {
	return c.options.ProxyProtoHeaders
}

// ProxyProtoHttps returns the proxy protocol header HTTPS values.
func (c *Config) ProxyProtoHttps() []string {
	return c.options.ProxyProtoHttps
}

// ProxyProtoHeaders returns a map with the proxy https protocol headers.
func (c *Config) ProxyProtoHeaders() map[string]string {
	p := len(c.options.ProxyProtoHeaders)
	h := make(map[string]string, p+1)

	if p == 0 {
		h[header.ForwardedProto] = header.ProtoHttps
		return h
	}

	for k, v := range c.options.ProxyProtoHeaders {
		if l := len(c.options.ProxyProtoHttps); l == 0 {
			h[v] = header.ProtoHttps
		} else if l > k {
			h[v] = c.options.ProxyProtoHttps[k]
		} else {
			h[v] = c.options.ProxyProtoHttps[0]
		}
	}

	return h
}

// HttpMode returns the server mode.
func (c *Config) HttpMode() string {
	if c.Prod() {
		return HttpModeProd
	} else if c.options.HttpMode == "" {
		if c.Debug() {
			return HttpModeDebug
		}

		return HttpModeProd
	}

	return c.options.HttpMode
}

// HttpCompression returns the http compression method (gzip, or none).
func (c *Config) HttpCompression() string {
	return strings.ToLower(strings.TrimSpace(c.options.HttpCompression))
}

// HttpCacheMaxAge returns the time in seconds until cached content expires.
func (c *Config) HttpCacheMaxAge() thumb.MaxAge {
	if c.options.HttpCacheMaxAge < 1 || c.options.HttpCacheMaxAge > 31536000 {
		// Default to one month.
		return thumb.CacheMaxAge
	}

	return thumb.MaxAge(c.options.HttpCacheMaxAge)
}

// HttpCachePublic checks whether static content may be cached by a CDN or caching proxy.
func (c *Config) HttpCachePublic() bool {
	if c.options.HttpCachePublic {
		return true
	}

	return c.options.CdnUrl != ""
}

// HttpHost returns the built-in HTTP server host name or IP address (empty for all interfaces).
func (c *Config) HttpHost() string {
	// when unix socket used as host, make host as default value. or http client will act weirdly.
	if c.options.HttpHost == "" || c.HttpHostAsSocketPath() != "" {
		return "0.0.0.0"
	}

	return c.options.HttpHost
}

// HttpPort returns the HTTP server port number.
func (c *Config) HttpPort() int {
	if c.options.HttpPort == 0 {
		return 2342
	}

	return c.options.HttpPort
}

// HttpHostAsSocketPath tries to parse the HttpHost as unix socket path. If failed, return empty string.
func (c *Config) HttpHostAsSocketPath() string {
	host := c.options.HttpHost
	if strings.HasPrefix(host, "unix:") && strings.Contains(host, "/") {
		return strings.TrimPrefix(host, "unix:")
	}
	return ""
}

// TemplatesPath returns the server templates path.
func (c *Config) TemplatesPath() string {
	return filepath.Join(c.AssetsPath(), "templates")
}

// CustomTemplatesPath returns the path to custom templates.
func (c *Config) CustomTemplatesPath() string {
	if dir := c.CustomAssetsPath(); dir == "" {
		return ""
	} else if dir = filepath.Join(dir, "templates"); fs.PathExists(dir) {
		return dir
	}

	return ""
}

// TemplateFiles returns the file paths of all templates found.
func (c *Config) TemplateFiles() []string {
	results := make([]string, 0, 32)

	var tmplPaths []string

	// Path set for custom templates?
	if cDir := c.CustomTemplatesPath(); cDir != "" {
		tmplPaths = []string{c.TemplatesPath(), cDir}
	} else {
		tmplPaths = []string{c.TemplatesPath()}
	}

	// Find template files.
	for _, dir := range tmplPaths {
		if dir == "" {
			continue
		}

		matches, err := filepath.Glob(regexp.QuoteMeta(dir) + "/[A-Za-z0-9]*.*")

		if err != nil {
			continue
		}

		for _, tmplName := range matches {
			results = append(results, tmplName)
		}
	}

	return results
}

// TemplateExists checks if a template with the given name exists (e.g. index.gohtml).
func (c *Config) TemplateExists(name string) bool {
	if found := fs.FileExists(filepath.Join(c.TemplatesPath(), name)); found {
		return true
	} else if dir := c.CustomTemplatesPath(); dir != "" {
		return fs.FileExists(filepath.Join(dir, name))
	} else {
		return false
	}
}

// TemplateName returns the name of the default template (e.g. index.gohtml).
func (c *Config) TemplateName() string {
	if s := c.Settings(); s != nil {
		if c.TemplateExists(s.Templates.Default) {
			return s.Templates.Default
		}
	}

	return "index.gohtml"
}

// StaticPath returns the static assets' path.
func (c *Config) StaticPath() string {
	return filepath.Join(c.AssetsPath(), "static")
}

// StaticFile returns the path to a static file.
func (c *Config) StaticFile(fileName string) string {
	return filepath.Join(c.AssetsPath(), "static", fileName)
}

// BuildPath returns the static build path.
func (c *Config) BuildPath() string {
	return filepath.Join(c.StaticPath(), "build")
}

// ImgPath returns the path to static image files.
func (c *Config) ImgPath() string {
	return filepath.Join(c.StaticPath(), "img")
}
