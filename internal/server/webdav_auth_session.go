package server

import (
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// WebDAVAuthSession returns the client session that belongs to the auth token provided, or returns nil if it was not found.
func WebDAVAuthSession(c *gin.Context, authToken string) (sess *entity.Session, user *entity.User, sid string, cached bool) {
	// Check if an auth token in a valid format was provided.
	if authToken == "" {
		// Return if no token was provided.
		return nil, nil, "", false
	} else if !rnd.IsAuthAny(authToken) {
		// Return if token does not match any of the supported formats.
		return nil, nil, "", false
	}

	// Get client IP address.
	clientIp := header.ClientIP(c)

	// Check failure rate limit and return nil if it has been exceeded.
	if limiter.Auth.Reject(clientIp) {
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

	// Count error towards failure rate limit, emits audit event, and returns nil?
	if sess == nil || err != nil {
		limiter.Auth.Reserve(clientIp)
		event.AuditErr([]string{header.ClientIP(c), "webdav", "access with invalid auth token", authn.Denied})
		return nil, nil, sid, false
	}

	// Update client IP and user agent of the session from the HTTP request context.
	sess.UpdateContext(c)

	// Return session and user.
	return sess, sess.User(), sid, false
}
