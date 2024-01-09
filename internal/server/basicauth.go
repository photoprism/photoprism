package server

import (
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	gc "github.com/patrickmn/go-cache"

	"github.com/photoprism/photoprism/internal/api"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/header"
)

// To improve performance, we use a basic auth cache
// with an expiration time of about 5 minutes.
var basicAuthExpiration = 5 * time.Minute
var basicAuthCache = gc.New(basicAuthExpiration, basicAuthExpiration)
var basicAuthMutex = sync.Mutex{}
var BasicAuthRealm = "Basic realm=\"WebDAV Authorization Required\""

// WebDAVAuth checks authentication and authentication
// before WebDAV requests are processed.
func WebDAVAuth(conf *config.Config) gin.HandlerFunc {
	// Helper function that extracts the login information from the request headers.
	var basicAuth = func(c *gin.Context) (username, password, cacheKey string, authorized bool) {
		// Extract credentials from the HTTP request headers.
		username, password, cacheKey = header.BasicAuth(c)

		// Fail if the username or password is empty, as
		// this is not allowed under any circumstances.
		if username == "" || password == "" || cacheKey == "" {
			return "", "", "", false
		}

		// To improve performance, check the cache for already authorized users.
		if user, found := basicAuthCache.Get(cacheKey); found && user != nil {
			// Add cached user information to the request context.
			c.Set(gin.AuthUserKey, user.(*entity.User))
			// Credentials have already been authorized within the configured
			// expiration time of the basic auth cache (about 5 minutes).
			return username, password, cacheKey, true
		} else {
			// Credentials found, but not pre-authorized. If successful, the
			// authorization will be cached for the next request.
			return username, password, cacheKey, false
		}
	}

	// Authentication handler that is called before WebDAV requests are processed.
	return func(c *gin.Context) {
		if c == nil {
			return
		}

		// Get basic authentication credentials.
		username, password, cacheKey, authorized := basicAuth(c)

		// Allow requests from already authorized users to be processed.
		if authorized {
			return
		}

		// Re-request authentication if credentials are missing or incomplete.
		if cacheKey == "" {
			c.Header("WWW-Authenticate", BasicAuthRealm)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Get the client IP address from the request headers
		// for use in logs and to enforce request rate limits.
		clientIp := api.ClientIP(c)

		// Check the authentication request rate to block the client after
		// too many failed attempts (10/req per minute by default).
		if limiter.Login.Reject(clientIp) {
			limiter.Abort(c)
			return
		}

		basicAuthMutex.Lock()
		defer basicAuthMutex.Unlock()

		// User credentials.
		f := form.Login{
			UserName: username,
			Password: password,
		}

		// Check credentials and authorization.
		if user, _, err := entity.Auth(f, nil, c); err != nil {
			message := err.Error()
			limiter.Login.Reserve(clientIp)
			event.AuditErr([]string{clientIp, "webdav login as %s", message}, clean.LogQuote(username))
			event.LoginError(clientIp, "webdav", username, api.UserAgent(c), message)
		} else if user == nil {
			message := "account not found"
			limiter.Login.Reserve(clientIp)
			event.AuditErr([]string{clientIp, "webdav login as %s", message}, clean.LogQuote(username))
			event.LoginError(clientIp, "webdav", username, api.UserAgent(c), message)
		} else if !user.CanUseWebDAV() {
			// Sync disabled for this account.
			message := "sync disabled"

			event.AuditWarn([]string{clientIp, "webdav login as %s", message}, clean.LogQuote(username))
			event.LoginError(clientIp, "webdav", username, api.UserAgent(c), message)
		} else if err = os.MkdirAll(filepath.Join(conf.OriginalsPath(), user.GetUploadPath()), fs.ModeDir); err != nil {
			message := "failed to create user upload path"

			event.AuditWarn([]string{clientIp, "webdav login as %s", message}, clean.LogQuote(username))
			event.LoginError(clientIp, "webdav", username, api.UserAgent(c), message)
		} else {
			// Successfully authenticated.
			event.AuditInfo([]string{clientIp, "webdav login as %s", "succeeded"}, clean.LogQuote(username))
			event.LoginInfo(clientIp, "webdav", username, api.UserAgent(c))

			// Cache successful authentication.
			basicAuthCache.SetDefault(cacheKey, user)
			c.Set(gin.AuthUserKey, user)
			return
		}

		// Abort request.
		c.Header("WWW-Authenticate", BasicAuthRealm)
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}
