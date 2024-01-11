package server

import (
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// WebDAVAuthSession returns the client session that belongs to the auth token provided, or returns nil if it was not found.
func WebDAVAuthSession(c *gin.Context, authToken string) (sess *entity.Session, user *entity.User, sid string, cached bool) {
	if authToken == "" {
		// Abort authentication if no token was provided.
		return nil, nil, "", false
	} else if !rnd.IsAuthToken(authToken) && !rnd.IsAuthSecret(authToken) {
		// Abort authentication if token doesn't match expected format.
		return nil, nil, "", false
	}

	// Get session ID for the auth token provided.
	sid = rnd.SessionID(authToken)

	// Check if client authorization has been cached to improve performance.
	if cacheData, found := webdavAuthCache.Get(sid); found && cacheData != nil {
		// Add cached user information to the request context.
		user = cacheData.(*entity.User)
		return nil, user, sid, true
	}

	var err error

	// Find the session based on the hashed token used as session ID and return it.
	sess, err = entity.FindSession(sid)

	// Log error and return nil if no matching session was found.
	if sess == nil || err != nil {
		event.AuditErr([]string{header.ClientIP(c), "access webdav", "invalid auth token or secret"})
		return nil, nil, sid, false
	}

	// Update the client IP and the user agent from
	// the request context if they have changed.
	sess.UpdateContext(c)

	// Returns session and user if all checks have passed.
	return sess, sess.User(), sid, false
}
