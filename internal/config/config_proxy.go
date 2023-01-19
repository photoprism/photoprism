package config

import (
	"os"
)

// HttpsProxy returns the HTTPS proxy to use for outgoing connections.
func (c *Config) HttpsProxy() string {
	if c.options.HttpsProxy != "" {
		return c.options.HttpsProxy
	} else if httpsProxy := os.Getenv("HTTPS_PROXY"); httpsProxy != "" {
		return httpsProxy
	}

	return ""
}

// HttpsProxyInsecure checks if invalid TLS certificates should be ignored when using the configured HTTPS proxy.
func (c *Config) HttpsProxyInsecure() bool {
	if c.HttpsProxy() == "" {
		return false
	}

	return c.options.HttpsProxyInsecure
}
