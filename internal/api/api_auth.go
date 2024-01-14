package api

import (
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/header"
)

// Auth checks if the user is authorized to access a resource with the given permission
// and returns the session or nil otherwise.
func Auth(c *gin.Context, resource acl.Resource, grant acl.Permission) *entity.Session {
	return AuthAny(c, resource, acl.Permissions{grant})
}

// AuthAny checks if the user is authorized to access a resource with any of the specified permissions
// and returns the session or nil otherwise.
func AuthAny(c *gin.Context, resource acl.Resource, grants acl.Permissions) (s *entity.Session) {
	// Get client IP and auth token from the request headers.
	clientIp := ClientIP(c)
	authToken := AuthToken(c)

	// Find active session to perform authorization check or deny if no session was found.
	if s = Session(clientIp, authToken); s == nil {
		event.AuditWarn([]string{clientIp, "unauthenticated", "%s %s", "denied"}, grants.String(), string(resource))
		return entity.SessionStatusUnauthorized()
	} else {
		s.SetClientIP(clientIp)
	}

	// If the request is from a client application, check its authorization based
	// on the allowed scope, the ACL, and the user account it belongs to (if any).
	if s.IsClient() {
		// Check ACL resource name against the permitted scope.
		if !s.HasScope(resource.String()) {
			event.AuditErr([]string{clientIp, "client %s", "session %s", "access %s", "denied"}, s.AuthID, s.RefID, string(resource))
			return s
		}

		// Perform an authorization check based on the ACL defaults for client applications.
		if acl.Resources.DenyAll(resource, acl.RoleClient, grants) {
			event.AuditErr([]string{clientIp, "client %s", "session %s", "%s %s", "denied"}, s.AuthID, s.RefID, grants.String(), string(resource))
			return entity.SessionStatusForbidden()
		}

		// Additionally check the user authorization if the client belongs to a user account.
		if s.NoUser() {
			// Allow access based on the ACL defaults for client applications.
			event.AuditInfo([]string{clientIp, "client %s", "session %s", "%s %s", "granted"}, s.AuthID, s.RefID, grants.String(), string(resource))
		} else if u := s.User(); !u.IsDisabled() && !u.IsUnknown() && u.IsRegistered() {
			if acl.Resources.DenyAll(resource, u.AclRole(), grants) {
				event.AuditErr([]string{clientIp, "client %s", "session %s", "%s %s as %s", "denied"}, s.AuthID, s.RefID, grants.String(), string(resource), u.String())
				return entity.SessionStatusForbidden()
			}

			// Allow access based on the user role.
			event.AuditInfo([]string{clientIp, "client %s", "session %s", "%s %s as %s", "granted"}, s.AuthID, s.RefID, grants.String(), string(resource), u.String())
		} else {
			// Deny access if it is not a regular user account or the account has been disabled.
			event.AuditErr([]string{clientIp, "client %s", "session %s", "%s %s as unauthorized user", "denied"}, s.AuthID, s.RefID, grants.String(), string(resource))
			return entity.SessionStatusForbidden()
		}

		return s
	}

	// Otherwise, perform a regular ACL authorization check based on the user role.
	if u := s.User(); u.IsUnknown() || u.IsDisabled() {
		event.AuditWarn([]string{clientIp, "session %s", "%s %s as unauthorized user", "denied"}, s.RefID, grants.String(), string(resource))
		return entity.SessionStatusUnauthorized()
	} else if acl.Resources.DenyAll(resource, u.AclRole(), grants) {
		event.AuditErr([]string{clientIp, "session %s", "%s %s as %s", "denied"}, s.RefID, grants.String(), string(resource), u.AclRole().String())
		return entity.SessionStatusForbidden()
	} else {
		event.AuditInfo([]string{clientIp, "session %s", "%s %s as %s", "granted"}, s.RefID, grants.String(), string(resource), u.AclRole().String())
		return s
	}
}

// AuthToken returns the client authentication token from the request context if one was found,
// or an empty string if no supported request header value was provided.
func AuthToken(c *gin.Context) string {
	return header.AuthToken(c)
}
