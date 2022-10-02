package api

import (
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/service"
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

		var sess *entity.Session
		var isNew bool

		// Find existing session, if any.
		if s := Session(SessionID(c)); s != nil {
			// Update existing session.
			sess = s
		} else {
			// Create new session.
			sess = service.Session().New(c)
			isNew = true
		}

		// Sign in and save session.
		if err := sess.SignIn(f, c); err != nil {
			c.AbortWithStatusJSON(sess.HttpStatus(), gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
			return
		} else if sess, err = service.Session().Save(sess); err != nil {
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

		// Get config values for use by the JavaScript UI and other clients.
		var clientConfig config.ClientConfig
		if conf := service.Config(); sess.User().IsVisitor() {
			clientConfig = conf.ClientShare()
		} else if sess.User().IsRegistered() {
			clientConfig = conf.ClientSession(sess)
		} else {
			clientConfig = conf.ClientPublic()
		}

		// Send JSON response with user information, session data, and client config values.
		c.JSON(sess.HttpStatus(), gin.H{"status": "ok", "id": sess.ID, "user": sess.User(), "data": sess.Data(), "config": clientConfig})
	})
}
