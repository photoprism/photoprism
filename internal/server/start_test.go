package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/crypto/acme/autocert"

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

func TestNewHTTPServer(t *testing.T) {
	t.Run("UsesConfiguredValues", func(t *testing.T) {
		conf := config.NewConfig(config.CliTestContext())
		conf.Options().HttpHeaderTimeout = 15 * time.Second
		conf.Options().HttpHeaderBytes = 2048
		conf.Options().HttpIdleTimeout = 2 * time.Minute

		server := newHTTPServer(http.NewServeMux(), conf)

		assert.Equal(t, 15*time.Second, server.ReadHeaderTimeout)
		assert.Equal(t, 0*time.Second, server.ReadTimeout)
		assert.Equal(t, 0*time.Second, server.WriteTimeout)
		assert.Equal(t, 2*time.Minute, server.IdleTimeout)
		assert.Equal(t, 2048, server.MaxHeaderBytes)
	})
	t.Run("UsesDefaultsWhenConfigIsNil", func(t *testing.T) {
		server := newHTTPServer(http.NewServeMux(), nil)

		assert.Equal(t, config.DefaultHttpHeaderTimeout, server.ReadHeaderTimeout)
		assert.Equal(t, 0*time.Second, server.ReadTimeout)
		assert.Equal(t, 0*time.Second, server.WriteTimeout)
		assert.Equal(t, config.DefaultHttpIdleTimeout, server.IdleTimeout)
		assert.Equal(t, config.DefaultHttpHeaderBytes, server.MaxHeaderBytes)
	})
}

func TestCanonicalRedirectTarget(t *testing.T) {
	t.Run("UsesConfiguredSiteHostInsteadOfRequestHost", func(t *testing.T) {
		conf := config.NewConfig(config.CliTestContext())
		conf.Options().SiteUrl = "https://photos.example.com:7443/library/"

		req := httptest.NewRequest(http.MethodGet, "http://evil.example.test/library/login?next=%2Flibrary", nil)
		req.Host = "evil.example.test"

		target := canonicalRedirectTarget(req, conf)

		assert.Equal(t, "https://photos.example.com:7443/library/login?next=%2Flibrary", target)
	})
	t.Run("FallsBackToRequestHostWithoutConfig", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://localhost:2342/library", nil)
		req.Host = "localhost:2342"

		target := canonicalRedirectTarget(req, nil)

		assert.Equal(t, "https://localhost:2342/library", target)
	})
}

func TestAutoTLSHTTPHandler(t *testing.T) {
	t.Run("UsesCanonicalHostForFallbackRedirect", func(t *testing.T) {
		conf := config.NewConfig(config.CliTestContext())
		conf.Options().SiteUrl = "https://photos.example.com:7443/library/"

		handler := (&autocert.Manager{
			HostPolicy: autocert.HostWhitelist(conf.SiteDomain()),
		}).HTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			redirect(w, r, conf)
		}))

		req := httptest.NewRequest(http.MethodGet, "http://evil.example.test/library/login?next=%2Flibrary", nil)
		req.Host = "evil.example.test"

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, httpsRedirect, w.Code)
		assert.Equal(t, "https://photos.example.com:7443/library/login?next=%2Flibrary", w.Header().Get(header.Location))
	})
	t.Run("HandlesAcmeChallengeWithoutFallbackRedirect", func(t *testing.T) {
		conf := config.NewConfig(config.CliTestContext())
		conf.Options().SiteUrl = "https://photos.example.com/"

		fallbackCalled := false
		handler := (&autocert.Manager{
			HostPolicy: autocert.HostWhitelist(conf.SiteDomain()),
		}).HTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fallbackCalled = true
			redirect(w, r, conf)
		}))

		req := httptest.NewRequest(http.MethodGet, "http://photos.example.com/.well-known/acme-challenge/test-token", nil)
		req.Host = conf.SiteDomain()

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.False(t, fallbackCalled)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Empty(t, w.Header().Get(header.Location))
	})
}
