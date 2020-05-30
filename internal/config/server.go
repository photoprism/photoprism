package config

import (
	"path/filepath"

	"github.com/photoprism/photoprism/pkg/fs"
)

// DetachServer returns true if server should detach from console (daemon mode).
func (c *Config) DetachServer() bool {
	return c.params.DetachServer
}

// HttpServerHost returns the built-in HTTP server host name or IP address (empty for all interfaces).
func (c *Config) HttpServerHost() string {
	if c.params.HttpServerHost == "" {
		return "0.0.0.0"
	}

	return c.params.HttpServerHost
}

// HttpServerPort returns the built-in HTTP server port.
func (c *Config) HttpServerPort() int {
	if c.params.HttpServerPort == 0 {
		return 2342
	}

	return c.params.HttpServerPort
}

// HttpServerMode returns the server mode.
func (c *Config) HttpServerMode() string {
	if c.params.HttpServerMode == "" {
		if c.Debug() {
			return "debug"
		}

		return "release"
	}

	return c.params.HttpServerMode
}

// HttpServerPassword returns the password for the user interface (optional).
func (c *Config) HttpServerPassword() string {
	return c.params.HttpServerPassword
}

// HttpTemplatesPath returns the server templates path.
func (c *Config) HttpTemplatesPath() string {
	return filepath.Join(c.ResourcesPath(), "templates")
}

// HttpTemplateExists returns true if a template with the given name exists (e.g. index.tmpl).
func (c *Config) HttpTemplateExists(name string) bool {
	return fs.FileExists(filepath.Join(c.HttpTemplatesPath(), name))
}

// HttpDefaultTemplate returns the name of the default template (e.g. index.tmpl).
func (c *Config) HttpDefaultTemplate() string {
	if c.HttpTemplateExists(c.Settings().Templates.Default) {
		return c.Settings().Templates.Default
	}

	return "index.tmpl"
}

// HttpFaviconsPath returns the favicons path.
func (c *Config) HttpFaviconsPath() string {
	return filepath.Join(c.HttpStaticPath(), "favicons")
}

// HttpStaticPath returns the static server assets path (//server/static/*).
func (c *Config) HttpStaticPath() string {
	return filepath.Join(c.ResourcesPath(), "static")
}

// HttpStaticBuildPath returns the static build path (//server/static/build/*).
func (c *Config) HttpStaticBuildPath() string {
	return filepath.Join(c.HttpStaticPath(), "build")
}
