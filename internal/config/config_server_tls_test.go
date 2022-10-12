package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_CertsPath(t *testing.T) {
	c := NewConfig(CliTestContext())
	if dir := c.CertsPath(); dir == "" {
		t.Fatal("certs path is empty")
	} else if !strings.HasPrefix(dir, c.ConfigPath()) {
		t.Fatalf("unexpected certs path: %s", dir)
	}
}

func TestConfig_AutoTLS(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.AutoTLS())
	c.options.AutoTLS = "hello@example.com"
	assert.Equal(t, "hello@example.com", c.AutoTLS())
	c.options.AutoTLS = "hello"
	assert.Equal(t, "", c.AutoTLS())
	c.options.AutoTLS = ""
	assert.Equal(t, "", c.AutoTLS())
}

func TestConfig_TLS(t *testing.T) {
	c := NewConfig(CliTestContext())

	cert, key := c.TLS()

	assert.Equal(t, "", cert)
	assert.Equal(t, "", key)
}

func TestConfig_TLSKey(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.TLSKey())
}

func TestConfig_TLSCert(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.TLSCert())
}

func TestConfig_HttpsPort(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, -1, c.HttpsPort())
}

func TestConfig_HttpsRedirect(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, -1, c.HttpsRedirect())
}
