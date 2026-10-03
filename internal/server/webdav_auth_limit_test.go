package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestWebDAVAuth_AuthLimit checks the answer to a WebDAV request with an auth token.
func TestWebDAVAuth_AuthLimit(t *testing.T) {
	conf := config.TestConfig()
	webdavHandler := WebDAVAuth(conf)

	origLimit := limiter.Auth
	t.Cleanup(func() { limiter.Auth = origLimit })
	limiter.Auth = limiter.NewLimit(rate.Every(24*time.Hour), 1)

	request := func(clientIp, token string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/originals/", nil)
		c.Request.RemoteAddr = clientIp + ":1234"
		header.SetAuthorization(c.Request, token)
		entity.FlushSessionCache()
		webdavHandler(c)
		return w
	}

	t.Run("OverLimit", func(t *testing.T) {
		for !limiter.Auth.Reject("198.51.100.71") {
			limiter.Auth.Reserve("198.51.100.71")
		}
		for _, token := range []string{rnd.AuthToken(), entity.SessionFixtures.Pointer("alice_token").AuthToken()} {
			w := request("198.51.100.71", token)
			assert.Equal(t, http.StatusTooManyRequests, w.Code)
			assert.Empty(t, w.Header().Get("WWW-Authenticate"))
		}
	})
	t.Run("AppPasswordOverLimit", func(t *testing.T) {
		exhaust := "198.51.100.73"
		for !limiter.Auth.Reject(exhaust) {
			limiter.Auth.Reserve(exhaust)
		}
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/originals/", nil)
		c.Request.RemoteAddr = exhaust + ":1234"
		c.Request.SetBasicAuth("alice", rnd.AppPassword())
		entity.FlushSessionCache()
		webdavHandler(c)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Empty(t, w.Header().Get("WWW-Authenticate"))
	})
	t.Run("BasicAuthOverLimit", func(t *testing.T) {
		exhaust := "198.51.100.74"
		for !limiter.Auth.Reject(exhaust) {
			limiter.Auth.Reserve(exhaust)
		}
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/originals/", nil)
		c.Request.RemoteAddr = exhaust + ":1234"
		c.Request.SetBasicAuth("alice", "not-an-app-password")
		entity.FlushSessionCache()
		webdavHandler(c)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Empty(t, w.Header().Get("WWW-Authenticate"))
	})
	t.Run("CachedBasicAuthOverLimit", func(t *testing.T) {
		// Credentials authorized within the cache lifetime are not accepted while the client is over the limit.
		basicRequest := func(clientIp string) *httptest.ResponseRecorder {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodGet, "/originals/", nil)
			c.Request.RemoteAddr = clientIp + ":1234"
			c.Request.SetBasicAuth("alice", "Alice123!")
			webdavHandler(c)
			return w
		}
		entity.FlushSessionCache()
		assert.Equal(t, http.StatusOK, basicRequest("198.51.100.76").Code)
		for !limiter.Auth.Reject("198.51.100.76") {
			limiter.Auth.Reserve("198.51.100.76")
		}
		w := basicRequest("198.51.100.76")
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Empty(t, w.Header().Get("WWW-Authenticate"))
	})
	t.Run("LoginRequestLimit", func(t *testing.T) {
		// The login request limit answers uncached basic auth credentials without a credential prompt.
		origLogin := limiter.Login
		t.Cleanup(func() { limiter.Login = origLogin })
		limiter.Login = limiter.NewLimit(rate.Every(24*time.Hour), 1)
		for !limiter.Login.Reject("198.51.100.78") {
			limiter.Login.Reserve("198.51.100.78")
		}
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/originals/", nil)
		c.Request.RemoteAddr = "198.51.100.78:1234"
		c.Request.SetBasicAuth("alice", "not-an-app-password")
		entity.FlushSessionCache()
		webdavHandler(c)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Empty(t, w.Header().Get("WWW-Authenticate"))
	})
	t.Run("NoCredentialsOverLimit", func(t *testing.T) {
		exhaust := "198.51.100.77"
		for !limiter.Auth.Reject(exhaust) {
			limiter.Auth.Reserve(exhaust)
		}
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/originals/", nil)
		c.Request.RemoteAddr = exhaust + ":1234"
		webdavHandler(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Equal(t, BasicAuthRealm, w.Header().Get("WWW-Authenticate"))
	})
	t.Run("NonTokenHeaderOverLimit", func(t *testing.T) {
		exhaust := "198.51.100.75"
		for !limiter.Auth.Reject(exhaust) {
			limiter.Auth.Reserve(exhaust)
		}
		w := request(exhaust, "not a token")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Equal(t, BasicAuthRealm, w.Header().Get("WWW-Authenticate"))
	})
	t.Run("InvalidTokenUnderLimit", func(t *testing.T) {
		w := request("198.51.100.72", rnd.AuthToken())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Equal(t, BasicAuthRealm, w.Header().Get("WWW-Authenticate"))
	})
}
