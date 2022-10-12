package config

import (
	"path/filepath"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// CertsPath returns the path to the TLS certificates and keys.
func (c *Config) CertsPath() string {
	return filepath.Join(c.ConfigPath(), "certs")
}

// AutoTLS returns the email address for enabling automatic HTTPS via Let's Encrypt.
func (c *Config) AutoTLS() string {
	return clean.Email(c.options.AutoTLS)
}

// TLSKey returns the HTTPS private key filename.
func (c *Config) TLSKey() string {
	if c.options.TLSKey == "" {
		return ""
	} else if fs.FileExistsNotEmpty(c.options.TLSKey) {
		return c.options.TLSKey
	} else if fileName := filepath.Join(c.CertsPath(), c.options.TLSKey); fs.FileExistsNotEmpty(fileName) {
		return fileName
	}

	return ""
}

// TLSCert returns the HTTPS certificate filename.
func (c *Config) TLSCert() string {
	if c.options.TLSCert == "" {
		return ""
	} else if fs.FileExistsNotEmpty(c.options.TLSCert) {
		return c.options.TLSCert
	} else if fileName := filepath.Join(c.CertsPath(), c.options.TLSCert); fs.FileExistsNotEmpty(fileName) {
		return fileName
	}

	return ""
}

// TLS returns the HTTPS certificate and private key file name.
func (c *Config) TLS() (certFile, privateKey string) {
	certFile = c.TLSCert()
	privateKey = c.TLSKey()

	if c.options.TLSCert == "" || privateKey == "" {
		return "", ""
	}

	return certFile, privateKey
}

// HttpsPort returns the HTTPS server port number.
func (c *Config) HttpsPort() int {
	if !c.SiteHttps() {
		return -1
	}

	if c.options.HttpsPort == 0 {
		return 2443
	}

	return c.options.HttpsPort
}

// HttpsRedirect returns the HTTPS redirect status code.
func (c *Config) HttpsRedirect() int {
	if !c.SiteHttps() {
		return -1
	}

	if c.options.HttpsRedirect > 0 && c.options.HttpsRedirect < 300 && c.options.HttpsRedirect >= 400 {
		return 301
	}

	return c.options.HttpsRedirect
}
