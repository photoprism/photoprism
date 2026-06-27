package wellknown

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
)

func TestPortalOpenIDConfiguration(t *testing.T) {
	conf := config.TestConfig()

	t.Run("Shape", func(t *testing.T) {
		result := NewPortalOpenIDConfiguration(conf)
		require.IsType(t, &OpenIDConfiguration{}, result)

		// The OP authorize/token/userinfo endpoints live under the APIv1 path.
		assert.Equal(t, "http://localhost:2342/", result.Issuer)
		assert.Equal(t, "http://localhost:2342/api/v1/oauth/authorize", result.AuthorizationEndpoint)
		assert.Equal(t, "http://localhost:2342/api/v1/oauth/token", result.TokenEndpoint)
		assert.Equal(t, "http://localhost:2342/api/v1/oauth/userinfo", result.UserinfoEndpoint)
		assert.Equal(t, "http://localhost:2342/api/v1/oauth/logout", result.EndSessionEndpoint)
		assert.Equal(t, "http://localhost:2342/.well-known/jwks.json", result.JwksUri)
	})
	t.Run("Capabilities", func(t *testing.T) {
		result := NewPortalOpenIDConfiguration(conf)

		// All capability lists must match the values published in the spec —
		// instances key off the exact strings to know what to send.
		assert.Equal(t, []string{"code"}, result.ResponseTypesSupported)
		assert.Equal(t, []string{"authorization_code"}, result.GrantTypesSupported)
		assert.Equal(t, []string{"public"}, result.SubjectTypesSupported)
		assert.Equal(t, []string{"EdDSA"}, result.IdTokenSigningAlgValuesSupported)
		assert.Equal(t, []string{"openid", "profile", "email", "cluster", "groups"}, result.ScopesSupported)
		assert.Equal(t, []string{"S256"}, result.CodeChallengeMethodsSupported)
		assert.Equal(t, []string{"client_secret_basic", "client_secret_post"}, result.TokenEndpointAuthMethodsSupported)
	})
	t.Run("BasePathDeployment", func(t *testing.T) {
		// For a sub-path Portal (SiteUrl carries a base path), every advertised
		// URL must contain the base path exactly once — the issuer already carries
		// it, so the appended paths must be bare. Guards against the JwksUri
		// base-path doubling regression.
		c := config.NewConfig(config.CliTestContext())
		c.Options().SiteUrl = "http://foo:2342/foo/"
		result := NewPortalOpenIDConfiguration(c)

		for _, u := range []string{result.Issuer, result.AuthorizationEndpoint, result.TokenEndpoint, result.UserinfoEndpoint, result.EndSessionEndpoint, result.JwksUri} {
			assert.Equal(t, 1, strings.Count(u, "/foo/"), "base path must appear exactly once in %s", u)
			assert.NotContains(t, u, "/foo/foo/", "base path must not be doubled in %s", u)
		}
		assert.Equal(t, "http://foo:2342/foo/api/v1/oauth/authorize", result.AuthorizationEndpoint)
		assert.Equal(t, "http://foo:2342/foo/api/v1/oauth/token", result.TokenEndpoint)
		assert.Equal(t, "http://foo:2342/foo/api/v1/oauth/userinfo", result.UserinfoEndpoint)
		assert.Equal(t, "http://foo:2342/foo/api/v1/oauth/logout", result.EndSessionEndpoint)
		assert.Equal(t, "http://foo:2342/foo/.well-known/jwks.json", result.JwksUri)
	})
	t.Run("IssuerWithTrailingSlashIsNormalized", func(t *testing.T) {
		// The Portal-issuer accessor falls through to SiteUrl which already
		// includes a trailing slash; the generator must not double it.
		result := NewPortalOpenIDConfiguration(conf)
		assert.Equal(t, "http://localhost:2342/", result.Issuer, "issuer must end with exactly one slash")
		assert.NotContains(t, result.AuthorizationEndpoint, "//api")
		assert.NotContains(t, result.TokenEndpoint, "//api")
		assert.NotContains(t, result.UserinfoEndpoint, "//api")
	})
}
