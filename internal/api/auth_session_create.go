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
)

// CreateSession creates a new client session and returns it as JSON if authentication was successful.
//
// POST /api/v1/session
func CreateSession(router *gin.RouterGroup) {
	router.POST("/session", func(c *gin.Context) {
		var f form.Login

		if err := c.BindJSON(&f); err != nil {
			event.AuditWarn([]string{ClientIP(c), "create session", "invalid request", "%s"}, err)
			AbortBadRequest(c)
			return
		}

		conf := get.Config()

		// Skip authentication if app is running in public mode.
		if conf.Public() {
			sess := get.Session().Public()
			data := gin.H{
				"status":   "ok",
				"id":       sess.ID,
				"provider": sess.AuthProvider,
				"user":     sess.User(),
				"data":     sess.Data(),
				"config":   conf.ClientPublic(),
			}
			c.JSON(http.StatusOK, data)
			return
		}

		// Check limit for failed auth requests (max. 10 per minute).
		if limiter.Login.Reject(ClientIP(c)) {
			limiter.AbortJSON(c)
			return
		}

		var sess *entity.Session
		var isNew bool

		// Find existing session, if any.
		if s := Session(SessionID(c)); s != nil {
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
			event.AuditErr([]string{ClientIP(c), "%s"}, err)
			c.AbortWithStatusJSON(sess.HttpStatus(), gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
			return
		} else if sess == nil {
			c.AbortWithStatusJSON(sess.HttpStatus(), gin.H{"error": i18n.Msg(i18n.ErrUnexpected)})
			return
		} else if isNew {
			event.AuditInfo([]string{ClientIP(c), "session %s", "created"}, sess.RefID)
		} else {
			event.AuditInfo([]string{ClientIP(c), "session %s", "updated"}, sess.RefID)
		}

		// Add session id to response headers.
		AddSessionHeader(c, sess.ID)

		// Get config values for the UI.
		clientConfig := conf.ClientSession(sess)

		// User information, session data, and client config values.
		data := gin.H{
			"status":   "ok",
			"id":       sess.ID,
			"provider": sess.AuthProvider,
			"user":     sess.User(),
			"data":     sess.Data(),
			"config":   clientConfig,
		}

		// Send JSON response.
		c.JSON(sess.HttpStatus(), data)
	})
}
