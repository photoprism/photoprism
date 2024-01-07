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

		// Check authentication token.
		if id == "" {
			// Abort if authentication token is missing or empty.
			AbortBadRequest(c)
			return
		}

		conf := get.Config()
		authToken := AuthToken(c)

		// Skip authentication if app is running in public mode.
		var sess *entity.Session
		if conf.Public() {
			sess = get.Session().Public()
			id = sess.ID
			authToken = sess.AuthToken()
		} else {
			sess = Session(authToken)
		}

		switch {
		case sess == nil:
			AbortUnauthorized(c)
			return
		case sess.Expired(), sess.ID == "":
			AbortUnauthorized(c)
			return
		case sess.Invalid(), sess.ID != id && !conf.Public():
			AbortForbidden(c)
			return
		}

		// Update user information.
		sess.RefreshUser()

		// Add session id to response headers.
		AddSessionHeader(c, authToken)

		// Response includes user data, session data, and client config values.
		response := GetSessionResponse(authToken, sess, get.Config().ClientSession(sess))

		// Return JSON response.
		c.JSON(http.StatusOK, response)
	})
}
