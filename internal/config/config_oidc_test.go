package config

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_OIDCEnabled(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.False(t, c.OIDCEnabled())
}

func TestConfig_OIDCIssuer(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.OIDCIssuer())
}

func TestConfig_OIDCIssuerURL(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.IsType(t, &url.URL{}, c.OIDCIssuerURL())
	assert.Equal(t, "", c.OIDCIssuerURL().Path)

	c.options.OIDCIssuer = "test"
	assert.Equal(t, "test", c.OIDCIssuerURL().Path)
}

func TestConfig_OIDCClient(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.OIDCClient())
}

func TestConfig_OIDCSecret(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.OIDCSecret())
}

func TestConfig_OIDCScopes(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, OIDCDefaultScopes, c.OIDCScopes())

	c.options.OIDCScopes = ""

	assert.Equal(t, OIDCDefaultScopes, c.OIDCScopes())
}

func TestConfig_OIDCRegister(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.False(t, c.OIDCRegister())
}

func TestConfig_OIDCInsecure(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.False(t, c.OIDCInsecure())
}

func TestConfig_OIDCReport(t *testing.T) {
	c := NewConfig(CliTestContext())

	r, _ := c.OIDCReport()
	assert.GreaterOrEqual(t, len(r), 6)
}
