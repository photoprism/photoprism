package config

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/pkg/authn"
)

func TestConfig_OIDCEnabled(t *testing.T) {
	t.Run("DisableForHTTP", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.SiteUrl = "http://photos.myphotos.com"
		assert.False(t, c.OIDCEnabled())
	})
	t.Run("OIDCDisabled", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.DisableOIDC = true
		assert.False(t, c.OIDCEnabled())
	})
	t.Run("InvalidOIDCUri", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.SiteUrl = "https://photos.myphotos.com"
		assert.True(t, c.SiteHttps())
		c.options.OIDCUri = "http://example.com"
		assert.False(t, c.OIDCEnabled())
	})
	t.Run("Enabled", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.SiteUrl = "https://photos.myphotos.com"
		assert.True(t, c.SiteHttps())
		c.options.OIDCUri = "https://example.com"
		c.options.OIDCClient = "test"
		c.options.OIDCSecret = "test123467"
		assert.True(t, c.OIDCEnabled())
	})
}

func TestConfig_OIDCUri(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.IsType(t, &url.URL{}, c.OIDCUri())
	assert.Equal(t, "", c.OIDCUri().Path)

	c.options.OIDCUri = "test"
	assert.Equal(t, "", c.OIDCUri().String())
	assert.Equal(t, "", c.OIDCUri().Path)

	c.options.OIDCUri = "http://test/"
	assert.Equal(t, "", c.OIDCUri().String())
	assert.Equal(t, "", c.OIDCUri().Path)

	c.options.OIDCUri = "https://test/"
	assert.Equal(t, "https://test/", c.OIDCUri().String())
	assert.Equal(t, "/", c.OIDCUri().Path)

	c.options.OIDCUri = ""
	assert.IsType(t, &url.URL{}, c.OIDCUri())
	assert.Equal(t, "", c.OIDCUri().String())
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

	assert.Equal(t, authn.OidcScopes, c.OIDCScopes())

	c.options.OIDCScopes = ""

	assert.Equal(t, authn.OidcScopes, c.OIDCScopes())
}

func TestConfig_OIDCProvider(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "OpenID", c.OIDCProvider())

	c.options.OIDCProvider = "test"

	assert.Equal(t, "test", c.OIDCProvider())
}

func TestConfig_OIDCIcon(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "/static/img/oidc.svg", c.OIDCIcon())

	c.options.OIDCIcon = "./test.svg"

	assert.Equal(t, "./test.svg", c.OIDCIcon())
}

func TestConfig_OIDCRedirect(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.False(t, c.OIDCRedirect())
}

func TestConfig_OIDCUsername(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, authn.ClaimPreferredUsername, c.OIDCUsername())

	c.options.OIDCUsername = "email"

	assert.Equal(t, authn.ClaimEmail, c.OIDCUsername())

	c.options.OIDCUsername = "name"

	assert.Equal(t, authn.ClaimName, c.OIDCUsername())

	c.options.OIDCUsername = "nickname"

	assert.Equal(t, authn.ClaimNickname, c.OIDCUsername())

	c.options.OIDCUsername = ""

	assert.Equal(t, authn.ClaimPreferredUsername, c.OIDCUsername())
}

func TestConfig_OIDCDomain(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.OIDCDomain())

	c.options.OIDCDomain = "example.com"

	assert.Equal(t, "example.com", c.OIDCDomain())

	c.options.OIDCDomain = "foo"

	assert.Equal(t, "", c.OIDCDomain())

	c.options.OIDCDomain = ""

	assert.Equal(t, "", c.OIDCDomain())
}

func TestConfig_OIDCRegister(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.False(t, c.OIDCRegister())
}

func TestConfig_OIDCRole(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, acl.RoleGuest, c.OIDCRole())

	c.options.OIDCRole = "invalid"

	assert.Equal(t, acl.RoleNone, c.OIDCRole())

	c.options.OIDCRole = "admin"

	assert.Equal(t, acl.RoleAdmin, c.OIDCRole())
}

func TestConfig_OIDCWebDAV(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.False(t, c.OIDCWebDAV())
}

func TestConfig_DisableOIDC(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.False(t, c.DisableOIDC())
}

func TestConfig_OIDCLoginUri(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "/api/v1/oidc/login", c.OIDCLoginUri())
}

func TestConfig_OIDCRedirectUri(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "/api/v1/oidc/redirect", c.OIDCRedirectUri())
}

func TestConfig_OIDCReport(t *testing.T) {
	c := NewConfig(CliTestContext())

	r, _ := c.OIDCReport()
	assert.GreaterOrEqual(t, len(r), 6)
}
