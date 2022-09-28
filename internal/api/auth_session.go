package api

import (
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/internal/session"
	"github.com/photoprism/photoprism/pkg/clean"
)

// SessionID returns the session ID from the request context.
func SessionID(c *gin.Context) (sessId string) {
	if c == nil {
		// Should never happen.
		return ""
	}

	// Get the authentication token from the HTTP headers.
	return clean.ID(c.GetHeader(session.Header))
}

// Session finds the client session for the given ID or returns nil otherwise.
func Session(id string) *entity.Session {
	// Return default session when public mode is enabled.
	if service.Config().Public() {
		return service.Session().Public()
	} else if id == "" {
		return nil
	}

	// Find session or otherwise return nil.
	s, err := service.Session().Get(id)

	if err != nil {
		return nil
	}

	return &s
}
