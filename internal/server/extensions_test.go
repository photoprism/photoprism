package server

import (
	"errors"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
)

// resetExtensions restores the registered extensions and the init guard after a test.
func resetExtensions(t *testing.T) {
	t.Helper()

	orig := Ext()
	extInit = sync.Once{}

	t.Cleanup(func() {
		extensions.Store(orig)
		extInit = sync.Once{}
	})
}

func TestRegister(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		resetExtensions(t)
		n := len(Ext())

		Register("zz-test", func(router *gin.Engine, conf *config.Config) error { return nil })

		ext := Ext()
		require.Len(t, ext, n+1)
		assert.Equal(t, "zz-test", ext[n].name)
		assert.NotNil(t, ext[n].init)
	})
}

func TestExt(t *testing.T) {
	t.Run("PreservesOrder", func(t *testing.T) {
		resetExtensions(t)
		n := len(Ext())

		Register("zz-first", func(router *gin.Engine, conf *config.Config) error { return nil })
		Register("zz-second", func(router *gin.Engine, conf *config.Config) error { return nil })

		ext := Ext()
		require.Len(t, ext, n+2)
		assert.Equal(t, "zz-first", ext[n].name)
		assert.Equal(t, "zz-second", ext[n+1].name)
	})
}

func TestExtensions_Init(t *testing.T) {
	t.Run("RunsOnceAndReportsErrors", func(t *testing.T) {
		resetExtensions(t)
		hook := captureRecoveryLog(t, logrus.TraceLevel)

		router := gin.New()
		conf := config.NewMinimalTestConfig(t.TempDir())

		var calls []string

		ext := Extensions{
			{name: "zz-ok", init: func(r *gin.Engine, c *config.Config) error {
				assert.Same(t, router, r)
				assert.Same(t, conf, c)
				calls = append(calls, "zz-ok")
				return nil
			}},
			{name: "zz-fail", init: func(r *gin.Engine, c *config.Config) error {
				calls = append(calls, "zz-fail")
				return errors.New("zz-error")
			}},
		}

		ext.Init(router, conf)
		ext.Init(router, conf)

		assert.Equal(t, []string{"zz-ok", "zz-fail"}, calls)

		var warned bool

		for _, entry := range hook.AllEntries() {
			if entry.Level == logrus.WarnLevel {
				warned = true
				assert.Contains(t, entry.Message, "zz-error when loading zz-fail extension")
			}
		}

		assert.True(t, warned)
	})
}
