package wellknown

import (
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

		// Endpoints follow the Portal OIDC OP path scheme (no /api/v1/ prefix).
		assert.Equal(t, "http://localhost:2342/", result.Issuer)
		assert.Equal(t, "http://localhost:2342/oauth/authorize", result.AuthorizationEndpoint)
		assert.Equal(t, "http://localhost:2342/oauth/token", result.TokenEndpoint)
		assert.Equal(t, "http://localhost:2342/oauth/userinfo", result.UserinfoEndpoint)
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

	t.Run("IssuerWithTrailingSlashIsNormalized", func(t *testing.T) {
		// The Portal-issuer accessor falls through to SiteUrl which already
		// includes a trailing slash; the generator must not double it.
		result := NewPortalOpenIDConfiguration(conf)
		assert.Equal(t, "http://localhost:2342/", result.Issuer, "issuer must end with exactly one slash")
		assert.NotContains(t, result.AuthorizationEndpoint, "//oauth")
		assert.NotContains(t, result.TokenEndpoint, "//oauth")
		assert.NotContains(t, result.UserinfoEndpoint, "//oauth")
	})
}
