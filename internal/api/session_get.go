package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/get"
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
		// Prevent CDNs from caching this endpoint.
		if header.IsCdn(c.Request) {
			AbortNotFound(c)
			return
		}

		id := clean.ID(c.Param("id"))

		if id != "" && !rnd.IsSessionID(id) {
			// Abort if session id is provided but invalid.
			AbortBadRequest(c)
			return
		}

		conf := get.Config()

		// Check if the session user is allowed to manage all accounts or update his/her own account.
		s := AuthAny(c, acl.ResourceSessions, acl.Permissions{acl.ActionManage, acl.ActionView})

		// Check if session is valid.
		switch {
		case s.Abort(c):
			return
		case s.Expired(), s.ID == "":
			AbortUnauthorized(c)
			return
		case s.Invalid(), id != "" && s.ID != id && !conf.Public():
			AbortForbidden(c)
			return
		}

		// Get auth token from headers.
		authToken := AuthToken(c)

		// Update user information.
		s.RefreshUser()

		// Response includes user data, session data, and client config values.
		response := GetSessionResponse(authToken, s, get.Config().ClientSession(s))

		// Return JSON response.
		c.JSON(http.StatusOK, response)
	}

	router.GET("/session", getSessionHandler)
	router.GET("/session/:id", getSessionHandler)
	router.GET("/sessions/:id", getSessionHandler)
}
