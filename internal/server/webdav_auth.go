package server

import (
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	gc "github.com/patrickmn/go-cache"

	"github.com/photoprism/photoprism/internal/api"
	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/authn"
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
			return username, password, "", false
		}

		// To improve performance, check the cache for already authorized users.
		if user, found := webdavAuthCache.Get(cacheKey); found && user != nil {
			// Add user to request context and return to signal successful authentication.
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

		// Add a vary response header for authentication, if any.
		if c.GetHeader(header.XAuthToken) != "" {
			c.Writer.Header().Add(header.Vary, header.XAuthToken)
		} else if c.GetHeader(header.XSessionID) != "" {
			c.Writer.Header().Add(header.Vary, header.XSessionID)
		}

		// Get basic authentication credentials, if any.
		username, password, cacheKey, authorized := basicAuth(c)

		// Allow requests from already authorized users to be processed.
		if authorized {
			return
		}

		// Get the client IP address from the request headers
		// for use in logs and to enforce request rate limits.
		clientIp := header.ClientIP(c)

		// Get access token, if any.
		authToken := header.AuthToken(c)

		// Use the value provided in the password field as auth token if no username was provided
		// and the format matches an app password e.g. "OXiV72-wTtiL9-d04jO7-X7XP4p".
		if username != "" && authToken == "" && rnd.IsAppPassword(password, true) {
			authToken = password
		}

		// Check webdav access authorization using an auth token or app password, if provided.
		if sess, user, sid, cached := WebDAVAuthSession(c, authToken); user != nil && cached {
			// Add user to request context to signal successful authentication if username is empty or matches.
			if username == "" || strings.EqualFold(clean.Username(username), user.Username()) {
				c.Set(gin.AuthUserKey, user)
				return
			}

			limiter.Auth.Reserve(clientIp)
			event.AuditErr([]string{clientIp, "webdav", "access as %s with authorization granted to %s", authn.Denied}, clean.Log(username), clean.Log(user.Username()))
			WebDAVAbortUnauthorized(c)
			return
		} else if sess == nil {
			// Ignore and try basic auth next.
		} else if !sess.HasUser() || user == nil {
			// Log error if session does not belong to an authorized user account.
			event.AuditErr([]string{clientIp, "webdav", "client %s", "session %s", "access without user account", authn.Denied}, clean.Log(sess.ClientInfo()), sess.RefID)
			WebDAVAbortUnauthorized(c)
			return
		} else if sess.IsClient() && sess.InsufficientScope(acl.ResourceWebDAV, nil) {
			// Log error if the client is allowed to access webdav based on its scope.
			message := authn.ErrInsufficientScope.Error()
			event.AuditWarn([]string{clientIp, "webdav", "client %s", "session %s", "access as %s", message}, clean.Log(sess.ClientInfo()), sess.RefID, clean.LogQuote(user.Username()))
			WebDAVAbortUnauthorized(c)
			return
		} else if !user.CanUseWebDAV() {
			// Log warning if WebDAV is disabled for this account.
			message := authn.ErrWebDAVAccessDisabled.Error()
			event.AuditWarn([]string{clientIp, "webdav", "client %s", "session %s", "access as %s", message}, clean.Log(sess.ClientInfo()), sess.RefID, clean.LogQuote(user.Username()))
			WebDAVAbortUnauthorized(c)
			return
		} else if username != "" && !strings.EqualFold(clean.Username(username), user.Username()) {
			limiter.Auth.Reserve(clientIp)
			// Log warning if auth token username and specified username do not match.
			message := authn.ErrUsernameDoesNotMatch.Error()
			event.AuditWarn([]string{clientIp, "webdav", "client %s", "session %s", "access as %s", message}, clean.Log(sess.ClientInfo()), sess.RefID, clean.LogQuote(user.Username()))
			WebDAVAbortUnauthorized(c)
			return
		} else if err := fs.MkdirAll(filepath.Join(conf.OriginalsPath(), user.GetUploadPath())); err != nil {
			// Log warning if upload path could not be created.
			message := authn.ErrFailedToCreateUploadPath.Error()
			event.AuditWarn([]string{clientIp, "webdav", "client %s", "session %s", "access as %s", message}, clean.Log(sess.ClientInfo()), sess.RefID, clean.LogQuote(user.Username()))
			WebDAVAbortServerError(c)
			return
		} else {
			// Update session activity.
			sess.UpdateLastActive(true)

			// Log successful authentication.
			event.AuditInfo([]string{clientIp, "webdav", "client %s", "session %s", "access as %s", authn.Succeeded}, clean.Log(sess.ClientInfo()), sess.RefID, clean.LogQuote(user.Username()))
			event.LoginInfo(clientIp, "webdav", user.Username(), api.UserAgent(c))

			// Cache authentication to improve performance.
			webdavAuthCache.SetDefault(sid, user)

			// Add user to request context and return to signal successful authentication.
			c.Set(gin.AuthUserKey, user)
			return
		}

		// Re-request authentication if credentials are missing or incomplete.
		if cacheKey == "" {
			WebDAVAbortUnauthorized(c)
			return
		}

		// Check request rate limit.
		r := limiter.Login.Request(clientIp)

		// Abort if request rate limit is exceeded.
		if r.Reject() || limiter.Auth.Reject(clientIp) {
			c.Header("WWW-Authenticate", BasicAuthRealm)
			limiter.Abort(c)
			return
		}

		webdavAuthMutex.Lock()
		defer webdavAuthMutex.Unlock()

		// User credentials.
		f := form.Login{
			Username: username,
			Password: password,
		}

		// Check credentials and authorization.
		if user, _, _, err := entity.Auth(f, nil, c); err != nil {
			// Abort if authentication has failed.
			message := authn.ErrInvalidCredentials.Error()
			event.AuditErr([]string{clientIp, "webdav", "login as %s", message}, clean.LogQuote(username))
			event.LoginError(clientIp, "webdav", username, api.UserAgent(c), message)
		} else if user == nil {
			// Abort if account was not found.
			message := authn.ErrAccountNotFound.Error()
			event.AuditErr([]string{clientIp, "webdav", "login as %s", message}, clean.LogQuote(username))
			event.LoginError(clientIp, "webdav", username, api.UserAgent(c), message)
		} else if !user.CanUseWebDAV() {
			// Return the reserved request rate limit tokens, even if account isn't allowed to use WebDAV.
			r.Success()

			// Abort if WebDAV is disabled for this account.
			message := authn.ErrWebDAVAccessDisabled.Error()
			event.AuditWarn([]string{clientIp, "webdav", "login as %s", message}, clean.LogQuote(username))
			event.LoginError(clientIp, "webdav", username, api.UserAgent(c), message)
		} else if err = fs.MkdirAll(filepath.Join(conf.OriginalsPath(), user.GetUploadPath())); err != nil {
			// Return the reserved request rate limit tokens, even if path could not be created.
			r.Success()

			// Abort if upload path could not be created.
			message := authn.ErrFailedToCreateUploadPath.Error()
			event.AuditWarn([]string{clientIp, "webdav", "login as %s", message}, clean.LogQuote(username))
			event.LoginError(clientIp, "webdav", username, api.UserAgent(c), message)
			WebDAVAbortServerError(c)
			return
		} else {
			// Return the reserved request rate limit tokens after successful authentication.
			r.Success()

			// Log successful authentication.
			event.AuditInfo([]string{clientIp, "webdav", "login as %s", authn.Succeeded}, clean.LogQuote(username))
			event.LoginInfo(clientIp, "webdav", username, api.UserAgent(c))

			// Cache authentication to improve performance.
			webdavAuthCache.SetDefault(cacheKey, user)

			// Add user to request context and return to signal successful authentication.
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

// WebDAVAbortServerError aborts the request with the status internal server error.
func WebDAVAbortServerError(c *gin.Context) {
	c.Header("WWW-Authenticate", BasicAuthRealm)
	c.AbortWithStatus(http.StatusInternalServerError)
}
