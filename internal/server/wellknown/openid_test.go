package wellknown

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
)

func TestOpenIDConfiguration(t *testing.T) {
	conf := config.TestConfig()

	t.Run("New", func(t *testing.T) {
		result := NewOpenIDConfiguration(conf)
		assert.IsType(t, &OpenIDConfiguration{}, result)
		assert.Equal(t, "http://localhost:2342/api/v1/oauth/token", result.TokenEndpoint)
		assert.Equal(t, "http://localhost:2342/api/v1/oauth/revoke", result.RevocationEndpoint)
		assert.Equal(t, "http://localhost:2342/.well-known/jwks.json", result.JwksUri)
		assert.Equal(t, OAuthResponseTypes, result.ResponseTypesSupported)
		assert.Equal(t, OAuthRevocationEndpointAuthMethods, result.RevocationEndpointAuthMethodsSupported)
	})
	t.Run("BasePathDeployment", func(t *testing.T) {
		// For a sub-path deployment, every advertised URL must contain the base
		// path exactly once. Guards against the JwksUri base-path doubling.
		c := config.NewConfig(config.CliTestContext())
		c.Options().SiteUrl = "http://foo:2342/foo/"
		result := NewOpenIDConfiguration(c)

		for _, u := range []string{result.Issuer, result.AuthorizationEndpoint, result.TokenEndpoint, result.UserinfoEndpoint, result.RevocationEndpoint, result.JwksUri} {
			assert.Equal(t, 1, strings.Count(u, "/foo/"), "base path must appear exactly once in %s", u)
			assert.NotContains(t, u, "/foo/foo/", "base path must not be doubled in %s", u)
		}
		assert.Equal(t, "http://foo:2342/foo/api/v1/oauth/token", result.TokenEndpoint)
		assert.Equal(t, "http://foo:2342/foo/.well-known/jwks.json", result.JwksUri)
	})
}
