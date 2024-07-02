package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/header"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// CreateSession creates a new client session and returns it as JSON if authentication was successful.
//
// POST /api/v1/session
// POST /api/v1/sessions
func CreateSession(router *gin.RouterGroup) {
	createSessionHandler := func(c *gin.Context) {
		// Prevent CDNs from caching this endpoint.
		if header.IsCdn(c.Request) {
			AbortNotFound(c)
			return
		}

		var f form.Login

		clientIp := ClientIP(c)

		// Assign and validate request form values.
		if err := c.BindJSON(&f); err != nil {
			event.AuditWarn([]string{clientIp, "create session", "invalid request", "%s"}, err)
			AbortBadRequest(c)
			return
		}

		// Disable caching of responses.
		c.Header(header.CacheControl, header.CacheControlNoStore)

		conf := get.Config()

		// Skip authentication if app is running in public mode.
		if conf.Public() {
			sess := get.Session().Public()

			// Response includes admin account data, session data, and client config values.
			response := CreateSessionResponse(sess.AuthToken(), sess, conf.ClientPublic())

			// Return JSON response.
			c.JSON(http.StatusOK, response)
			return
		}

		// Check request rate limit.
		var r *limiter.Request
		if f.HasPasscode() {
			r = limiter.Login.RequestN(clientIp, 3)
		} else {
			r = limiter.Login.Request(clientIp)
		}

		// Abort if failure rate limit is exceeded.
		if r.Reject() || limiter.Auth.Reject(clientIp) {
			limiter.AbortJSON(c)
			return
		}

		var sess *entity.Session
		var isNew bool
		var err error

		// Find existing session, if any.
		if s := Session(clientIp, AuthToken(c)); s != nil {
			// Update existing session.
			sess = s
		} else {
			// Create new session.
			sess = get.Session().New(c)
			isNew = true
		}

		// Check authentication credentials.
		if err = sess.LogIn(f, c); err != nil {
			if sess.Method().IsNot(authn.Method2FA) {
				c.AbortWithStatusJSON(sess.HttpStatus(), gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
			} else if errors.Is(err, authn.ErrPasscodeRequired) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error(), "code": i18n.ErrPasscodeRequired, "message": i18n.Msg(i18n.ErrPasscodeRequired)})
				// Return the reserved request rate limit tokens if password is correct, even if the verification code is missing.
				r.Success()
			} else {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error(), "code": i18n.ErrInvalidPasscode, "message": i18n.Msg(i18n.ErrInvalidPasscode)})
			}
			return
		}

		// Extend session lifetime if 2-Factor Authentication (2FA) is enabled for the account.
		if sess.Is2FA() && !sess.IsClient() {
			sess.SetExpiresIn(conf.SessionMaxAge() * 2)
			sess.SetTimeout(conf.SessionTimeout() * 2)
		}

		// Save session after successful authentication.
		if sess, err = get.Session().Save(sess); err != nil {
			event.AuditErr([]string{clientIp, "%s"}, err)
			c.AbortWithStatusJSON(sess.HttpStatus(), gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
			return
		} else if sess == nil {
			c.AbortWithStatusJSON(sess.HttpStatus(), gin.H{"error": i18n.Msg(i18n.ErrUnexpected)})
			return
		} else if isNew {
			event.AuditInfo([]string{clientIp, "session %s", "created"}, sess.RefID)
		} else {
			event.AuditInfo([]string{clientIp, "session %s", "updated"}, sess.RefID)
		}

		// Return the reserved request rate limit tokens after successful authentication.
		r.Success()

		// Response includes user data, session data, and client config values.
		response := CreateSessionResponse(sess.AuthToken(), sess, conf.ClientSession(sess))

		// Return JSON response.
		c.JSON(sess.HttpStatus(), response)
	}

	router.POST("/session", createSessionHandler)
	router.POST("/sessions", createSessionHandler)
}
