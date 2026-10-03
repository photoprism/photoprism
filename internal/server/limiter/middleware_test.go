package limiter

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"

	"github.com/photoprism/photoprism/pkg/http/header"
)

func TestMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// newRouter returns a router with the request limiter and the given trusted platform header.
	newRouter := func(t *testing.T, platform string) *gin.Engine {
		t.Helper()
		t.Cleanup(func() { header.SetTrustedPlatform("") })

		router := gin.New()
		router.TrustedPlatform = header.SetTrustedPlatform(platform)
		require.NoError(t, router.SetTrustedProxies(nil))
		router.Use(Middleware(NewLimit(rate.Every(time.Hour), 3)))
		router.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

		return router
	}

	// request sends a request from the peer with the given X-Forwarded-For lines and returns the status.
	request := func(router *gin.Engine, forwardedFor ...string) int {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.128.2.4:1234"

		for _, v := range forwardedFor {
			req.Header.Add(header.XForwardedFor, v)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		return w.Code
	}

	t.Run("PlatformAddress", func(t *testing.T) {
		router := newRouter(t, header.XForwardedFor)

		for i := 0; i < 3; i++ {
			assert.Equal(t, http.StatusOK, request(router, "198.51.100.7", "203.0.113.5"))
		}

		assert.Equal(t, http.StatusTooManyRequests, request(router, "198.51.100.8", "203.0.113.5"))
		assert.Equal(t, http.StatusTooManyRequests, request(router, "2001:0db8:0000:0000:0000:0000:0000:0001, 203.0.113.5"))
		assert.Equal(t, http.StatusOK, request(router, "203.0.113.5", "203.0.113.6"))
	})
	t.Run("IPv6Network", func(t *testing.T) {
		router := newRouter(t, "")

		// requestFrom sends a request from the specified peer and returns the status.
		requestFrom := func(remoteAddr string) int {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = remoteAddr
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			return w.Code
		}

		for i := range 3 {
			assert.Equal(t, http.StatusOK, requestFrom(fmt.Sprintf("[2001:db8:1:2::%x]:1234", i+1)))
		}

		assert.Equal(t, http.StatusTooManyRequests, requestFrom("[2001:db8:1:2::ff]:1234"))
		assert.Equal(t, http.StatusOK, requestFrom("[2001:db8:1:3::1]:1234"))
	})
	t.Run("PeerAddress", func(t *testing.T) {
		router := newRouter(t, "")

		for i := 0; i < 3; i++ {
			assert.Equal(t, http.StatusOK, request(router, "203.0.113.5"))
		}

		assert.Equal(t, http.StatusTooManyRequests, request(router, "203.0.113.6"))
	})
}
