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
