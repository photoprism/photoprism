package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
)

func TestRegisterWebDAVRoutes(t *testing.T) {
	// newConf returns a minimal config with authentication and existing originals and import folders,
	// since WebDAV is not available in public mode.
	newConf := func(t *testing.T) *config.Config {
		t.Helper()

		dir := t.TempDir()
		conf := config.NewMinimalTestConfig(dir)
		conf.Options().AuthMode = config.AuthModePasswd
		conf.Options().Public = false
		require.False(t, conf.DisableWebDAV())
		conf.Options().OriginalsPath = filepath.Join(dir, "originals")
		conf.Options().ImportPath = filepath.Join(dir, "import")
		require.NoError(t, os.MkdirAll(conf.OriginalsPath(), fs.ModeDir))
		require.NoError(t, os.MkdirAll(conf.ImportPath(), fs.ModeDir))

		return conf
	}

	t.Run("Enabled", func(t *testing.T) {
		conf := newConf(t)
		router := gin.New()
		registerWebDAVRoutes(router, conf)

		assert.Positive(t, countRoutesWithPrefix(router, conf.BaseUri(WebDAVOriginals)))
		assert.Positive(t, countRoutesWithPrefix(router, conf.BaseUri(WebDAVImport)))

		methods := make(map[string]bool)

		for _, r := range router.Routes() {
			methods[r.Method] = true
		}

		assert.True(t, methods[header.MethodPropfind])
		assert.True(t, methods[header.MethodPut])
	})
	t.Run("ReadOnly", func(t *testing.T) {
		conf := newConf(t)
		conf.Options().ReadOnly = true
		router := gin.New()
		registerWebDAVRoutes(router, conf)

		assert.Positive(t, countRoutesWithPrefix(router, conf.BaseUri(WebDAVOriginals)))
	})
	t.Run("Disabled", func(t *testing.T) {
		conf := newConf(t)
		conf.Options().DisableWebDAV = true
		router := gin.New()
		registerWebDAVRoutes(router, conf)

		assert.Empty(t, router.Routes())
	})
}
