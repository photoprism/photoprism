package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	t.Run("LogsRequest", func(t *testing.T) {
		hook := captureRecoveryLog(t, logrus.DebugLevel)

		router := gin.New()
		router.Use(Logger())
		router.GET("/teapot", func(c *gin.Context) {
			c.Status(http.StatusTeapot)
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/teapot?q=1", nil))

		assert.Equal(t, http.StatusTeapot, w.Code)
		require.Len(t, hook.AllEntries(), 1)

		entry := hook.LastEntry()
		assert.Equal(t, logrus.DebugLevel, entry.Level)
		assert.Contains(t, entry.Message, "server: GET /teapot?q=1 (418)")
	})
	t.Run("SilentAboveDebugLevel", func(t *testing.T) {
		hook := captureRecoveryLog(t, logrus.InfoLevel)

		router := gin.New()
		router.Use(Logger())
		router.GET("/ok", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ok", nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, hook.AllEntries())
	})
}
