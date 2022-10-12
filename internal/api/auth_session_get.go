package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/clean"
)

// GetSession returns the session data as JSON if authentication was successful.
//
// GET /api/v1/session/:id
func GetSession(router *gin.RouterGroup) {
	router.GET("/session/:id", func(c *gin.Context) {
		conf := service.Config()

		// Skip authentication if app is running in public mode.
		if conf.Public() {
			sess := service.Session().Public()
			c.JSON(http.StatusOK, gin.H{"status": "ok", "id": sess.ID, "user": sess.User(), "data": sess.Data(), "config": conf.ClientPublic()})
			return
		}

		id := clean.ID(c.Param("id"))

		if id == "" {
			AbortBadRequest(c)
			return
		} else if id != SessionID(c) {
			AbortForbidden(c)
			return
		}

		sess := Session(id)

		switch {
		case sess == nil:
			AbortUnauthorized(c)
			return
		case sess.Expired(), sess.ID == "":
			AbortUnauthorized(c)
			return
		case sess.Invalid():
			AbortForbidden(c)
			return
		}

		// Update user information.
		sess.RefreshUser()

		// Add session id to response headers.
		AddSessionHeader(c, sess.ID)

		var clientConfig config.ClientConfig

		if conf == nil {
			log.Errorf("session: config is not set - possible bug")
			AbortUnexpected(c)
			return
		} else if sess.User().IsVisitor() {
			clientConfig = conf.ClientShare()
		} else if sess.User().IsRegistered() {
			clientConfig = conf.ClientSession(sess)
		} else {
			clientConfig = conf.ClientPublic()
		}

		// Send JSON response with user information, session data, and client config values.
		c.JSON(http.StatusOK, gin.H{"status": "ok", "id": sess.ID, "user": sess.User(), "data": sess.Data(), "config": clientConfig})
	})
}
