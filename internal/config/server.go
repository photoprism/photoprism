package config

import "github.com/photoprism/photoprism/pkg/fs"

// DatabasePath returns the database storage path for TiDB.
func (c *Config) DatabasePath() string {
	if c.params.DatabasePath == "" {
		return c.ResourcesPath() + "/database"
	}

	return fs.Abs(c.params.DatabasePath)
}

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
	return c.ResourcesPath() + "/templates"
}

// HttpFaviconsPath returns the favicons path.
func (c *Config) HttpFaviconsPath() string {
	return c.HttpStaticPath() + "/favicons"
}

// HttpStaticPath returns the static server assets path (//server/static/*).
func (c *Config) HttpStaticPath() string {
	return c.ResourcesPath() + "/static"
}

// HttpStaticBuildPath returns the static build path (//server/static/build/*).
func (c *Config) HttpStaticBuildPath() string {
	return c.HttpStaticPath() + "/build"
}

// SqlServerHost returns the built-in SQL server host name or IP address (empty for all interfaces).
func (c *Config) SqlServerHost() string {
	if c.params.SqlServerHost == "" {
		return "127.0.0.1"
	}

	return c.params.SqlServerHost
}

// SqlServerPort returns the built-in SQL server port.
func (c *Config) SqlServerPort() uint {
	if c.params.SqlServerPort == 0 {
		return 4000
	}

	return c.params.SqlServerPort
}

// SqlServerPassword returns the password for the built-in database server.
func (c *Config) SqlServerPassword() string {
	return c.params.SqlServerPassword
}
