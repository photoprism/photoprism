package wellknown

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
)

func TestOAuthAuthorizationServer(t *testing.T) {
	conf := config.TestConfig()

	t.Run("New", func(t *testing.T) {
		result := NewOAuthAuthorizationServer(conf)
		assert.IsType(t, &OAuthAuthorizationServer{}, result)
		assert.Equal(t, "http://localhost:2342/api/v1/oauth/token", result.TokenEndpoint)
		assert.Equal(t, "http://localhost:2342/api/v1/oauth/revoke", result.RevocationEndpoint)
		assert.Equal(t, OAuthResponseTypes, result.ResponseTypesSupported)
		assert.Equal(t, OAuthRevocationEndpointAuthMethods, result.RevocationEndpointAuthMethodsSupported)
	})
}

func TestOAuthAuthorizationServer_AppFlow(t *testing.T) {
	conf := config.TestConfig()
	opt := conf.Options()
	apps, public, authMode := opt.OAuthApps, opt.Public, opt.AuthMode
	t.Cleanup(func() { opt.OAuthApps, opt.Public, opt.AuthMode = apps, public, authMode })
	opt.Public, opt.AuthMode = false, config.AuthModePasswd

	t.Run("Enabled", func(t *testing.T) {
		opt.OAuthApps = config.OAuthAppsRegistered
		result := NewOAuthAuthorizationServer(conf)
		assert.Equal(t, []string{"token", "code"}, result.ResponseTypesSupported)
		assert.Equal(t, []string{"client_credentials", "authorization_code", "refresh_token"}, result.GrantTypesSupported)
		assert.Equal(t, []string{"client_secret_basic", "client_secret_post", "none"}, result.TokenEndpointAuthMethodsSupported)
		assert.Equal(t, []string{"S256"}, result.CodeChallengeMethodsSupported)
		assert.True(t, result.AuthorizationResponseIssParameterSupported)
		assert.False(t, result.ClientIdMetadataDocumentSupported)
		assert.Equal(t, []string{"token"}, OAuthResponseTypes)
	})
	t.Run("Metadata", func(t *testing.T) {
		opt.OAuthApps = config.OAuthAppsAny
		assert.True(t, NewOAuthAuthorizationServer(conf).ClientIdMetadataDocumentSupported)
		assert.True(t, NewOpenIDConfiguration(conf).ClientIdMetadataDocumentSupported)

		opt.OAuthApps = "https://claude.ai"
		assert.True(t, NewOAuthAuthorizationServer(conf).ClientIdMetadataDocumentSupported)
	})
	t.Run("Disabled", func(t *testing.T) {
		opt.OAuthApps = config.OAuthAppsNone
		result := NewOAuthAuthorizationServer(conf)
		assert.Equal(t, OAuthResponseTypes, result.ResponseTypesSupported)
		assert.Equal(t, OAuthGrantTypes, result.GrantTypesSupported)
		assert.Equal(t, OAuthTokenEndpointAuthMethods, result.TokenEndpointAuthMethodsSupported)
		assert.Empty(t, result.CodeChallengeMethodsSupported)
		assert.False(t, result.AuthorizationResponseIssParameterSupported)
	})
}

func TestOAuthValues(t *testing.T) {
	conf := config.TestConfig()
	opt := conf.Options()
	apps, public, authMode := opt.OAuthApps, opt.Public, opt.AuthMode
	t.Cleanup(func() { opt.OAuthApps, opt.Public, opt.AuthMode = apps, public, authMode })
	opt.Public, opt.AuthMode = false, config.AuthModePasswd

	t.Run("Enabled", func(t *testing.T) {
		opt.OAuthApps = config.OAuthAppsRegistered
		values := []string{"a"}
		assert.Equal(t, []string{"a", "b"}, oauthValues(conf, values, []string{"b"}))
		assert.Equal(t, []string{"a"}, values)
	})
	t.Run("Disabled", func(t *testing.T) {
		opt.OAuthApps = config.OAuthAppsNone
		assert.Equal(t, []string{"a"}, oauthValues(conf, []string{"a"}, []string{"b"}))
	})
}
