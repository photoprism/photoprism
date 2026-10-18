package wellknown

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
)

func TestNewOAuthProtectedResource(t *testing.T) {
	conf := config.TestConfig()

	t.Run("Default", func(t *testing.T) {
		result := NewOAuthProtectedResource(conf)
		assert.Equal(t, "http://localhost:2342/api/v1/mcp", result.Resource)
		assert.Equal(t, []string{"http://localhost:2342/"}, result.AuthorizationServers)
		assert.Equal(t, []string{"mcp"}, result.ScopesSupported)
		assert.Equal(t, []string{"header"}, result.BearerMethodsSupported)
		assert.Equal(t, conf.SiteTitle(), result.ResourceName)
	})
}
