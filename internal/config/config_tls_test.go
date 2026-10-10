package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_CertificatesPath(t *testing.T) {
	c := NewConfig(CliTestContext())
	if dir := c.CertificatesPath(); dir == "" {
		t.Fatal("certificates path is empty")
	} else if !strings.HasPrefix(dir, c.ConfigPath()) {
		t.Fatalf("unexpected certificates path: %s", dir)
	}
}

func TestConfig_TLSEmail(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.TLSEmail())
	c.options.TLSEmail = "hello@example.com"
	assert.Equal(t, "hello@example.com", c.TLSEmail())
	c.options.TLSEmail = "hello"
	assert.Equal(t, "", c.TLSEmail())
	c.options.TLSEmail = ""
	assert.Equal(t, "", c.TLSEmail())
}

func TestConfig_TLSCert(t *testing.T) {
	c := NewConfig(CliTestContext())

	// Remember original values.
	defaultTls := c.options.DefaultTLS
	disableTls := c.options.DisableTLS

	c.options.DefaultTLS = false
	c.options.DisableTLS = true
	assert.Equal(t, "", c.TLSCert())
	assert.Equal(t, "", c.TLSKey())
	c.options.DisableTLS = false
	assert.Equal(t, "", c.TLSCert())
	assert.Equal(t, "", c.TLSKey())
	c.options.DefaultTLS = true
	assert.NotEmpty(t, c.TLSCert())
	assert.NotEmpty(t, c.TLSKey())
	assert.True(t, strings.HasSuffix(c.TLSCert(), "photoprism.crt"))
	assert.True(t, strings.HasSuffix(c.TLSKey(), "photoprism.key"))
	c.options.DefaultTLS = false
	assert.Equal(t, "", c.TLSCert())
	assert.Equal(t, "", c.TLSKey())

	// Restore original values.
	c.options.DefaultTLS = defaultTls
	c.options.DisableTLS = disableTls
}

func TestConfig_TLS(t *testing.T) {
	c := NewConfig(CliTestContext())

	cert, key := c.TLS()

	assert.Equal(t, "", cert)
	assert.Equal(t, "", key)
}

func TestConfig_AutoTLS(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		assert.False(t, NewConfig(CliTestContext()).AutoTLS())
	})
	t.Run("Enabled", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.SiteUrl = "https://photos.example.com/"
		c.options.TLSEmail = "admin@example.com"
		assert.True(t, c.AutoTLS())
	})
	t.Run("LongLabelAndIDN", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.TLSEmail = "admin@example.com"
		c.options.SiteUrl = "https://photoprism-family-archive-berlin-01.example.org/"
		assert.True(t, c.AutoTLS())
		c.options.SiteUrl = "https://bücher.example.org/"
		assert.True(t, c.AutoTLS())
	})
	t.Run("Disabled", func(t *testing.T) {
		for name, update := range map[string]func(c *Config){
			"DisableTLS":   func(c *Config) { c.options.DisableTLS = true },
			"NoEmail":      func(c *Config) { c.options.TLSEmail = "" },
			"InvalidEmail": func(c *Config) { c.options.TLSEmail = "admin" },
			"HttpSiteUrl":  func(c *Config) { c.options.SiteUrl = "http://photos.example.com/" },
			"IPv4":         func(c *Config) { c.options.SiteUrl = "https://192.0.2.1/" },
			"IPv6":         func(c *Config) { c.options.SiteUrl = "https://[2001:db8::1]/" },
			"Localhost":    func(c *Config) { c.options.SiteUrl = "https://localhost/" },
			"LocalDomain":  func(c *Config) { c.options.SiteUrl = "https://photos.local/" },
			"UnixSocket":   func(c *Config) { c.options.HttpHost = "unix:/tmp/photoprism.sock" },
		} {
			t.Run(name, func(t *testing.T) {
				c := NewConfig(CliTestContext())
				c.options.SiteUrl = "https://photos.example.com/"
				c.options.TLSEmail = "admin@example.com"
				update(c)
				assert.False(t, c.AutoTLS())
			})
		}
	})
}

func TestConfig_DisableTLS(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		assert.True(t, NewConfig(CliTestContext()).DisableTLS())
	})
	t.Run("AutoTLS", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.SiteUrl = "https://photos.example.com/"
		c.options.TLSEmail = "admin@example.com"
		assert.False(t, c.DisableTLS())
		c.options.DisableTLS = true
		assert.True(t, c.DisableTLS())
	})
}

func TestConfig_DefaultTLS(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.False(t, c.DefaultTLS())
}
