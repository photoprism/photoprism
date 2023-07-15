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

	c.options.DefaultTLS = false
	assert.Equal(t, "", c.TLSCert())
	c.options.DefaultTLS = true
	assert.True(t, strings.HasSuffix(c.TLSCert(), "photoprism.crt"))
	c.options.DefaultTLS = false
	assert.Equal(t, "", c.TLSCert())
}

func TestConfig_TLSKey(t *testing.T) {
	c := NewConfig(CliTestContext())

	c.options.DefaultTLS = false
	assert.Equal(t, "", c.TLSKey())
	c.options.DefaultTLS = true
	assert.True(t, strings.HasSuffix(c.TLSKey(), "photoprism.key"))
	c.options.DefaultTLS = false
	assert.Equal(t, "", c.TLSKey())
}

func TestConfig_TLS(t *testing.T) {
	c := NewConfig(CliTestContext())

	cert, key := c.TLS()

	assert.Equal(t, "", cert)
	assert.Equal(t, "", key)
}

func TestConfig_DisableTLS(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.True(t, c.DisableTLS())
}

func TestConfig_DefaultTLS(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.False(t, c.DefaultTLS())
}
