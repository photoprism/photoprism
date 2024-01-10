package server

import (
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	gc "github.com/patrickmn/go-cache"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/api"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Use auth cache to improve WebDAV performance. It has a standard expiration time of about 5 minutes.
var webdavAuthExpiration = 5 * time.Minute
var webdavAuthCache = gc.New(webdavAuthExpiration, webdavAuthExpiration)
var webdavAuthMutex = sync.Mutex{}
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
		if user, found := webdavAuthCache.Get(cacheKey); found && user != nil {
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

		// Get basic authentication credentials, if any.
		username, password, cacheKey, authorized := basicAuth(c)

		// Allow requests from already authorized users to be processed.
		if authorized {
			return
		}

		// Get the client IP address from the request headers
		// for use in logs and to enforce request rate limits.
		clientIp := api.ClientIP(c)

		// Get access token, if any.
		authToken := header.AuthToken(c)

		// Use the value provided in the password field as auth secret if no username was provided
		// and the format matches.
		if username == "" && authToken == "" && rnd.IsAuthSecret(password) {
			authToken = password
		}

		// Find client session if an auth token has been provided and perform authorization check.
		if authToken != "" {
			sid := rnd.SessionID(authToken)

			// Check if client authorization has been cached to improve performance.
			if user, found := webdavAuthCache.Get(sid); found && user != nil {
				// Add cached user information to the request context.
				c.Set(gin.AuthUserKey, user.(*entity.User))
				return
			}

			sess, err := entity.FindSession(sid)

			if sess == nil {
				limiter.Login.Reserve(clientIp)
				event.AuditErr([]string{clientIp, "access webdav", "invalid auth token"})
				WebDAVAbortUnauthorized(c)
				return
			} else if err != nil {
				limiter.Login.Reserve(clientIp)
				event.AuditErr([]string{clientIp, "access webdav", "%s"}, err.Error())
				WebDAVAbortUnauthorized(c)
				return
			} else {
				sess.UpdateContext(c)
			}

			// Required resource scope.
			resource := acl.ResourceWebDAV

			// If the request is from a client application, check its authorization based
			// on the allowed scope, the ACL, and the user account it belongs to (if any).
			if sess.IsClient() {
				// Check if client belongs to a user and if the "webdav" scope is set.
				if !sess.HasScope(resource.String()) || !sess.HasUser() {
					event.AuditErr([]string{clientIp, "client %s", "session %s", "access webdav", "denied"}, clean.Log(sess.AuthID), sess.RefID)
					WebDAVAbortUnauthorized(c)
					return
				}
			}

			// Check authorization and grant access if successful.
			if !sess.HasUser() {
				event.AuditErr([]string{clientIp, "session %s", "access webdav as unauthorized user", "denied"}, sess.RefID)
			} else if user := sess.User(); !user.CanUseWebDAV() {
				// Sync disabled for this account.
				message := "sync disabled"
				event.AuditWarn([]string{clientIp, "access webdav as %s", message}, clean.LogQuote(username))
			} else if err = os.MkdirAll(filepath.Join(conf.OriginalsPath(), user.GetUploadPath()), fs.ModeDir); err != nil {
				message := "failed to create user upload path"
				event.AuditWarn([]string{clientIp, "access webdav as %s", message}, clean.LogQuote(username))
			} else {
				// Cache successful authentication to improve performance.
				webdavAuthCache.SetDefault(sid, user)
				c.Set(gin.AuthUserKey, user)
				return
			}

			// Request authentication.
			WebDAVAbortUnauthorized(c)
			return
		}

		// Re-request authentication if credentials are missing or incomplete.
		if cacheKey == "" {
			WebDAVAbortUnauthorized(c)
			return
		}

		// Check the authentication request rate to block the client after
		// too many failed attempts (10/req per minute by default).
		if limiter.Login.Reject(clientIp) {
			limiter.Abort(c)
			return
		}

		webdavAuthMutex.Lock()
		defer webdavAuthMutex.Unlock()

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

			// Cache successful authentication to improve performance.
			webdavAuthCache.SetDefault(cacheKey, user)
			c.Set(gin.AuthUserKey, user)
			return
		}

		// Request authentication.
		WebDAVAbortUnauthorized(c)
	}
}

// WebDAVAbortUnauthorized aborts the request with the status unauthorized and requests authentication.
func WebDAVAbortUnauthorized(c *gin.Context) {
	c.Header("WWW-Authenticate", BasicAuthRealm)
	c.AbortWithStatus(http.StatusUnauthorized)
}
