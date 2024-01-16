package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/header"
)

// CreateSession creates a new client session and returns it as JSON if authentication was successful.
//
// POST /api/v1/session
// POST /api/v1/sessions
func CreateSession(router *gin.RouterGroup) {
	createSessionHandler := func(c *gin.Context) {
		// Disable caching of responses.
		c.Header(header.CacheControl, header.CacheControlNoStore)

		var f form.Login

		clientIp := ClientIP(c)

		if err := c.BindJSON(&f); err != nil {
			event.AuditWarn([]string{clientIp, "create session", "invalid request", "%s"}, err)
			AbortBadRequest(c)
			return
		}

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

		// Fail if authentication error rate limit is exceeded.
		if clientIp != "" && (limiter.Login.Reject(clientIp) || limiter.Auth.Reject(clientIp)) {
			limiter.AbortJSON(c)
			return
		}

		var sess *entity.Session
		var isNew bool

		// Find existing session, if any.
		if s := Session(clientIp, AuthToken(c)); s != nil {
			// Update existing session.
			sess = s
		} else {
			// Create new session.
			sess = get.Session().New(c)
			isNew = true
		}

		// Try to log in and save session if successful.
		if err := sess.LogIn(f, c); err != nil {
			c.AbortWithStatusJSON(sess.HttpStatus(), gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
			return
		} else if sess, err = get.Session().Save(sess); err != nil {
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

		// Response includes user data, session data, and client config values.
		response := CreateSessionResponse(sess.AuthToken(), sess, conf.ClientSession(sess))

		// Return JSON response.
		c.JSON(sess.HttpStatus(), response)
	}

	router.POST("/session", createSessionHandler)
	router.POST("/sessions", createSessionHandler)
}
