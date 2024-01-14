package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// GetSession returns the session data as JSON if authentication was successful.
//
// GET /api/v1/session
// GET /api/v1/session/:id
// GET /api/v1/sessions/:id
func GetSession(router *gin.RouterGroup) {
	getSessionHandler := func(c *gin.Context) {
		// Disable caching of responses.
		c.Header(header.CacheControl, header.CacheControlNoStore)

		id := clean.ID(c.Param("id"))

		// Abort if session id is provided but invalid.
		if id != "" && !rnd.IsSessionID(id) {
			AbortBadRequest(c)
			return
		}

		conf := get.Config()

		// Get client IP and auth token from request headers.
		clientIp := ClientIP(c)
		authToken := AuthToken(c)

		// Skip authentication if app is running in public mode.
		var sess *entity.Session
		if conf.Public() {
			sess = get.Session().Public()
			id = sess.ID
			authToken = sess.AuthToken()
		} else if clientIp != "" && limiter.Auth.Reject(clientIp) {
			// Fail if authentication error rate limit is exceeded.
			limiter.AbortJSON(c)
			return
		} else {
			sess = Session(clientIp, authToken)
		}

		switch {
		case sess == nil:
			if clientIp != "" {
				limiter.Auth.Reserve(clientIp)
			}
			AbortUnauthorized(c)
			return
		case sess.Expired(), sess.ID == "":
			AbortUnauthorized(c)
			return
		case sess.Invalid(), id != "" && sess.ID != id && !conf.Public():
			AbortForbidden(c)
			return
		}

		// Update user information.
		sess.RefreshUser()

		// Add auth token to response header.
		AddAuthTokenHeader(c, authToken)

		// Response includes user data, session data, and client config values.
		response := GetSessionResponse(authToken, sess, get.Config().ClientSession(sess))

		// Return JSON response.
		c.JSON(http.StatusOK, response)
	}

	router.GET("/session", getSessionHandler)
	router.GET("/session/:id", getSessionHandler)
	router.GET("/sessions/:id", getSessionHandler)
}
