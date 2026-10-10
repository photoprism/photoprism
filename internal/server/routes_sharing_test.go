package server

import (
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
)

// countRoutesWithPrefix returns the number of registered routes whose path starts with prefix.
func countRoutesWithPrefix(router *gin.Engine, prefix string) (n int) {
	for _, r := range router.Routes() {
		if strings.HasPrefix(r.Path, prefix) {
			n++
		}
	}

	return n
}

func TestRegisterSharingRoutes(t *testing.T) {
	t.Run("Enabled", func(t *testing.T) {
		conf := config.NewMinimalTestConfig(t.TempDir())
		router := gin.New()
		registerSharingRoutes(router, conf)

		routes := routeSet(router)
		assert.True(t, routes["GET "+conf.BaseUri("/s/:token")])
		assert.True(t, routes["GET "+conf.BaseUri("/s/:token/:shared")])
		assert.True(t, routes["GET "+conf.BaseUri("/s/:token/:shared/preview")])
	})
	t.Run("FrontendDisabled", func(t *testing.T) {
		conf := config.NewMinimalTestConfig(t.TempDir())
		conf.Options().DisableFrontend = true
		router := gin.New()
		registerSharingRoutes(router, conf)

		assert.Equal(t, 0, countRoutesWithPrefix(router, conf.BaseUri("/s/")))
	})
}
