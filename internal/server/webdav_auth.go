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
		// and the format matches auth secrets e.g. "iXrDz-aY16n-4IUWM-otkM3".
		if authToken == "" && rnd.IsAuthSecret(password) {
			authToken = password
		}

		// Allow webdav access based on the auth token or secret provided?
		if sess, user, sid, cached := WebDAVSession(c, authToken); cached && user != nil {
			// Add user to request context and return to signal successful authentication.
			c.Set(gin.AuthUserKey, user)
			return
		} else if sess == nil {
			// Ignore and try basic auth next.
		} else if !sess.HasUser() || user == nil {
			// Log error if session does not belong to an authorized user account.
			event.AuditErr([]string{clientIp, "session %s", "access webdav without authorized user account", "denied"}, sess.RefID)
		} else if sess.IsClient() && !sess.HasScope(acl.ResourceWebDAV.String()) {
			// Log error if the client is allowed to access webdav based on its scope.
			event.AuditErr([]string{clientIp, "client %s", "session %s", "access webdav without scope authorization", "denied"}, clean.Log(sess.AuthID), sess.RefID)
		} else if !user.CanUseWebDAV() {
			// Log warning if WebDAV is disabled for this account.
			message := "webdav access disabled"
			event.AuditWarn([]string{clientIp, "access webdav as %s", message}, clean.LogQuote(username))
		} else if err := os.MkdirAll(filepath.Join(conf.OriginalsPath(), user.GetUploadPath()), fs.ModeDir); err != nil {
			// Log warning if upload path could not be created.
			message := "failed to create user upload path"
			event.AuditWarn([]string{clientIp, "access webdav as %s", message}, clean.LogQuote(username))
		} else {
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
			// Abort if authentication has failed.
			message := err.Error()
			limiter.Login.Reserve(clientIp)
			event.AuditErr([]string{clientIp, "webdav login as %s", message}, clean.LogQuote(username))
			event.LoginError(clientIp, "webdav", username, api.UserAgent(c), message)
		} else if user == nil {
			// Abort if account was not found.
			message := "account not found"
			limiter.Login.Reserve(clientIp)
			event.AuditErr([]string{clientIp, "webdav login as %s", message}, clean.LogQuote(username))
			event.LoginError(clientIp, "webdav", username, api.UserAgent(c), message)
		} else if !user.CanUseWebDAV() {
			// Abort if WebDAV is disabled for this account.
			message := "webdav access disabled"
			event.AuditWarn([]string{clientIp, "webdav login as %s", message}, clean.LogQuote(username))
			event.LoginError(clientIp, "webdav", username, api.UserAgent(c), message)
		} else if err = os.MkdirAll(filepath.Join(conf.OriginalsPath(), user.GetUploadPath()), fs.ModeDir); err != nil {
			// Abort if upload path could not be created.
			message := "failed to create user upload path"
			event.AuditWarn([]string{clientIp, "webdav login as %s", message}, clean.LogQuote(username))
			event.LoginError(clientIp, "webdav", username, api.UserAgent(c), message)
		} else {
			// Log successful authentication.
			event.AuditInfo([]string{clientIp, "webdav login as %s", "succeeded"}, clean.LogQuote(username))
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

// WebDAVSession returns the client session that belongs to the auth token provided, or returns nil if it was not found.
func WebDAVSession(c *gin.Context, authToken string) (sess *entity.Session, user *entity.User, sid string, cached bool) {
	if authToken == "" {
		// Abort authentication if no token was provided.
		return nil, nil, "", false
	} else if !rnd.IsAuthToken(authToken) && !rnd.IsAuthSecret(authToken) {
		// Abort authentication if token doesn't match expected format.
		return nil, nil, "", false
	}

	// Get session ID for the auth token provided.
	sid = rnd.SessionID(authToken)

	// Check if client authorization has been cached to improve performance.
	if cacheData, found := webdavAuthCache.Get(sid); found && cacheData != nil {
		// Add cached user information to the request context.
		user = cacheData.(*entity.User)
		return nil, user, sid, true
	}

	var err error

	// Find the session based on the hashed token used as session ID and return it.
	sess, err = entity.FindSession(sid)

	// Log error and return nil if no matching session was found.
	if sess == nil || err != nil {
		event.AuditErr([]string{header.ClientIP(c), "access webdav", "invalid auth token or secret"})
		return nil, nil, sid, false
	}

	// Update the client IP and the user agent from
	// the request context if they have changed.
	sess.UpdateContext(c)

	// Returns session and user if all checks have passed.
	return sess, sess.User(), sid, false
}
