package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/log/status"
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

	var err error

	// Find the session based on the hashed token used as session ID and return it.
	sess, err = entity.FindSession(sid)

	// Count error towards failure rate limit, emits audit event, and returns nil?
	if sess == nil || err != nil {
		limiter.Auth.Reserve(clientIp)
		event.AuditErr([]string{header.ClientIP(c), "webdav", "access with invalid auth token", status.Denied})
		return nil, nil, sid, false
	}

	// Retain credential permissions when reusing the authenticated account.
	if cachedUser := entity.CachedWebDAVUser(sid); cachedUser != nil {
		return sess, cachedUser, sid, true
	}

	// Update client IP and user agent of the session from the HTTP request context.
	sess.UpdateContext(c)

	// Return session and user.
	return sess, sess.GetUser(), sid, false
}

// WebDAVSessionPermits checks the credential's WebDAV admission and requested action scope.
func WebDAVSessionPermits(sess *entity.Session, method string) bool {
	if sess == nil || !sess.Grants(acl.ResourceWebDAV, acl.ActionDownload) {
		return false
	}

	perms := WebDAVMethodPermissions(method)

	for _, perm := range perms {
		if !sess.Grants(acl.ResourceWebDAV, perm) {
			return false
		}
	}

	return !sess.IsClient() && !sess.HasScope() || sess.ValidateScope(acl.ResourceWebDAV, perms)
}

// WebDAVRequestPermits includes target-only property probes in upload scope.
func WebDAVRequestPermits(sess *entity.Session, r *http.Request) bool {
	if r == nil {
		return false
	}

	if WebDAVSessionPermits(sess, r.Method) {
		return true
	}

	depth := r.Header.Values("Depth")

	return r.Method == header.MethodPropfind && len(depth) == 1 && depth[0] == "0" &&
		WebDAVSessionPermits(sess, header.MethodPut)
}
