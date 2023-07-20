package config

import (
	"path/filepath"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

const (
	PrivateKeyExt = ".key"
	PublicCertExt = ".crt"
)

// CertificatesPath returns the path to the TLS certificates and keys.
func (c *Config) CertificatesPath() string {
	return filepath.Join(c.ConfigPath(), "certificates")
}

// TLSEmail returns the email address to enable automatic HTTPS via Let's Encrypt
func (c *Config) TLSEmail() string {
	return clean.Email(c.options.TLSEmail)
}

// TLSCert returns the public certificate required to enable TLS.
func (c *Config) TLSCert() string {
	certName := c.options.TLSCert
	if certName == "" {
		certName = c.SiteDomain() + PublicCertExt
	} else if fs.FileExistsNotEmpty(certName) {
		return certName
	}

	// find looks for an existing certificate file.
	find := func(certName string) string {
		if fileName := filepath.Join(c.CertificatesPath(), certName); fs.FileExistsNotEmpty(fileName) {
			return fileName
		} else if fileName = filepath.Join("/etc/ssl/certs", certName); fs.FileExistsNotEmpty(fileName) {
			return fileName
		} else {
			return ""
		}
	}

	// Find matching TLS certificate file.
	if fileName := find(certName); fileName != "" {
		return fileName
	}

	// Find default TLS certificate.
	if c.DefaultTLS() {
		return find("photoprism" + PublicCertExt)
	}

	// Not found.
	return ""
}

// TLSKey returns the private key required to enable TLS.
func (c *Config) TLSKey() string {
	keyName := c.options.TLSKey

	if keyName == "" {
		keyName = c.SiteDomain() + PrivateKeyExt
	} else if fs.FileExistsNotEmpty(keyName) {
		return keyName
	}

	// find looks for an existing private key file.
	find := func(keyName string) string {
		if fileName := filepath.Join(c.CertificatesPath(), keyName); fs.FileExistsNotEmpty(fileName) {
			return fileName
		} else if fileName = filepath.Join("/etc/ssl/private", keyName); fs.FileExistsNotEmpty(fileName) {
			return fileName
		} else {
			return ""
		}
	}

	// Find matching private key.
	if fileName := find(keyName); fileName != "" {
		return fileName
	}

	// Find default key file.
	if c.DefaultTLS() {
		return find("photoprism" + PrivateKeyExt)
	}

	// Not found.
	return ""
}

// TLS returns the HTTPS certificate and private key file name.
func (c *Config) TLS() (publicCert, privateKey string) {
	if c.DisableTLS() {
		return "", ""
	}

	return c.TLSCert(), c.TLSKey()
}

// DisableTLS checks if HTTPS should be disabled even if the site URL starts with https:// and a certificate is available.
func (c *Config) DisableTLS() bool {
	if c.options.DisableTLS {
		return true
	} else if !c.SiteHttps() {
		return true
	}

	return c.TLSCert() == "" || c.TLSKey() == ""
}

// DefaultTLS checks if a self-signed certificate should be used to enable HTTPS if no other certificate is available.
func (c *Config) DefaultTLS() bool {
	return c.options.DefaultTLS
}
