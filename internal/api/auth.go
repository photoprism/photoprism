package api

import (
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
)

// Auth checks if the user has permission to access the specified resource and returns the session if so.
func Auth(c *gin.Context, resource acl.Resource, grant acl.Permission) *entity.Session {
	return AuthAny(c, resource, acl.Permissions{grant})
}

// AuthAny checks if at least one permission allows access and returns the session in this case.
func AuthAny(c *gin.Context, resource acl.Resource, grants acl.Permissions) (s *entity.Session) {
	// Get the client IP and session ID from the request headers.
	ip := ClientIP(c)
	sid := SessionID(c)

	// Find active session to perform authorization check or deny if no session was found.
	if s = Session(sid); s == nil {
		event.AuditWarn([]string{ip, "unauthenticated", "%s %s", "denied"}, grants.String(), string(resource))
		return entity.SessionStatusUnauthorized()
	} else {
		s.SetClientIP(ip)
	}

	// If the request is from a client application, check its authorization based
	// on the allowed scope, the ACL, and the user account it belongs to (if any).
	if s.IsClient() {
		// Check ACL resource name against the permitted scope.
		if !s.HasScope(resource.String()) {
			event.AuditErr([]string{ip, "client %s", "session %s", "access %s", "denied"}, s.AuthID, s.RefID, string(resource))
			return s
		}

		// Perform an authorization check based on the ACL defaults for client applications.
		if acl.Resources.DenyAll(resource, acl.RoleClient, grants) {
			event.AuditErr([]string{ip, "client %s", "session %s", "%s %s", "denied"}, s.AuthID, s.RefID, grants.String(), string(resource))
			return entity.SessionStatusForbidden()
		}

		// Additionally check the user authorization if the client belongs to a user account.
		if s.NoUser() {
			// Allow access based on the ACL defaults for client applications.
			event.AuditInfo([]string{ip, "client %s", "session %s", "%s %s", "granted"}, s.AuthID, s.RefID, grants.String(), string(resource))
		} else if u := s.User(); !u.IsDisabled() && !u.IsUnknown() && u.IsRegistered() {
			if acl.Resources.DenyAll(resource, u.AclRole(), grants) {
				event.AuditErr([]string{ip, "client %s", "session %s", "%s %s as %s", "denied"}, s.AuthID, s.RefID, grants.String(), string(resource), u.String())
				return entity.SessionStatusForbidden()
			}

			// Allow access based on the user role.
			event.AuditInfo([]string{ip, "client %s", "session %s", "%s %s as %s", "granted"}, s.AuthID, s.RefID, grants.String(), string(resource), u.String())
		} else {
			// Deny access if it is not a regular user account or the account has been disabled.
			event.AuditErr([]string{ip, "client %s", "session %s", "%s %s as unauthorized user", "denied"}, s.AuthID, s.RefID, grants.String(), string(resource))
			return entity.SessionStatusForbidden()
		}

		return s
	}

	// Otherwise, perform a regular ACL authorization check based on the user role.
	if u := s.User(); u.IsUnknown() || u.IsDisabled() {
		event.AuditWarn([]string{ip, "session %s", "%s %s as unauthorized user", "denied"}, s.RefID, grants.String(), string(resource))
		return entity.SessionStatusUnauthorized()
	} else if acl.Resources.DenyAll(resource, u.AclRole(), grants) {
		event.AuditErr([]string{ip, "session %s", "%s %s as %s", "denied"}, s.RefID, grants.String(), string(resource), u.AclRole().String())
		return entity.SessionStatusForbidden()
	} else {
		event.AuditInfo([]string{ip, "session %s", "%s %s as %s", "granted"}, s.RefID, grants.String(), string(resource), u.AclRole().String())
		return s
	}
}
