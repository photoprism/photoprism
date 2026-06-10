package oidc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
)

func TestClientConfig(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		c := config.NewConfig(config.CliTestContext())
		result := ClientConfig(c, config.ClientPublic)
		assert.IsType(t, config.Values{}, result)
		assert.Equal(t, c.ClusterOIDC(), result["cluster"])
		assert.Equal(t, false, result["cluster"])
	})
	t.Run("ClusterOIDC", func(t *testing.T) {
		c := config.NewConfig(config.CliTestContext())
		c.Options().ClusterOIDC = true
		result := ClientConfig(c, config.ClientPublic)
		assert.Equal(t, true, result["cluster"])
	})
}
