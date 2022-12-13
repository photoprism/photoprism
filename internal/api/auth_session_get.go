package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/pkg/clean"
)

// GetSession returns the session data as JSON if authentication was successful.
//
// GET /api/v1/session/:id
func GetSession(router *gin.RouterGroup) {
	router.GET("/session/:id", func(c *gin.Context) {
		id := clean.ID(c.Param("id"))

		if id == "" {
			AbortBadRequest(c)
			return
		} else if id != SessionID(c) {
			AbortForbidden(c)
			return
		}

		conf := get.Config()

		// Skip authentication if app is running in public mode.
		var sess *entity.Session
		if conf.Public() {
			sess = get.Session().Public()
		} else {
			sess = Session(id)
		}

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

		// Send JSON response with user information, session data, and client config values.
		data := gin.H{
			"status":   "ok",
			"id":       sess.ID,
			"provider": sess.AuthProvider,
			"user":     sess.User(),
			"data":     sess.Data(),
			"config":   get.Config().ClientSession(sess),
		}

		c.JSON(http.StatusOK, data)
	})
}
