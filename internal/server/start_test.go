package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// newProxyTestRouter creates a test router with trusted proxy settings applied.
func newProxyTestRouter(conf *config.Config) *gin.Engine {
	r := gin.New()
	configureTrustedProxySettings(r, conf)

	r.GET("/ip", func(c *gin.Context) {
		c.String(http.StatusOK, header.ClientIP(c))
	})

	return r
}

// requestClientIP performs a test request and returns the resolved client IP.
func requestClientIP(t *testing.T, router *gin.Engine, remoteAddr, forwardedFor string) string {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/ip", nil)
	req.RemoteAddr = remoteAddr

	if forwardedFor != "" {
		req.Header.Set(header.XForwardedFor, forwardedFor)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	return w.Body.String()
}

func TestConfigureTrustedProxySettings(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("UsesForwardedIPForTrustedProxy", func(t *testing.T) {
		conf := config.NewConfig(config.CliTestContext())
		conf.Options().TrustedProxies = []string{header.CidrDockerInternal}
		conf.Options().ProxyClientHeaders = []string{header.XForwardedFor}

		router := newProxyTestRouter(conf)
		ip := requestClientIP(t, router, "172.16.5.10:12345", "203.0.113.9")

		assert.Equal(t, "203.0.113.9", ip)
	})
	t.Run("DisablesProxyTrustWhenNoTrustedProxiesConfigured", func(t *testing.T) {
		conf := config.NewConfig(config.CliTestContext())
		conf.Options().TrustedProxies = nil
		conf.Options().ProxyClientHeaders = []string{header.XForwardedFor}

		router := newProxyTestRouter(conf)
		ip := requestClientIP(t, router, "198.51.100.10:12345", "10.0.0.123")

		assert.Equal(t, "198.51.100.10", ip)
	})
	t.Run("FallsBackToDirectIPWhenTrustedProxyIsInvalid", func(t *testing.T) {
		conf := config.NewConfig(config.CliTestContext())
		conf.Options().TrustedProxies = []string{"invalid"}
		conf.Options().ProxyClientHeaders = []string{header.XForwardedFor}

		router := newProxyTestRouter(conf)
		ip := requestClientIP(t, router, "198.51.100.11:12345", "10.0.0.124")

		assert.Equal(t, "198.51.100.11", ip)
	})
}
