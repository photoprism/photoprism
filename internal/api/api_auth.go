package api

import (
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// Auth checks if the user is authorized to access a resource with the given permission
// and returns the session or nil otherwise.
func Auth(c *gin.Context, resource acl.Resource, perm acl.Permission) *entity.Session {
	return AuthAny(c, resource, acl.Permissions{perm})
}

// AuthAny checks if the user is authorized to access a resource with any of the specified permissions
// and returns the session or nil otherwise.
func AuthAny(c *gin.Context, resource acl.Resource, perms acl.Permissions) (s *entity.Session) {
	// Prevent CDNs from caching responses that require authentication.
	if header.IsCdn(c.Request) {
		return entity.SessionStatusForbidden()
	}

	// Get client IP and auth token from the request headers.
	clientIp := ClientIP(c)
	authToken := AuthToken(c)

	// Disable response caching.
	c.Header(header.CacheControl, header.CacheControlNoStore)

	// Allow requests based on an access token for specific resources.
	if resource == acl.ResourceVision && perms.Contains(acl.ActionUse) && vision.ServiceApi && vision.ServiceKey != "" && vision.ServiceKey == authToken {
		s = entity.NewSessionFromToken(c, authToken, acl.ResourceVision.String(), "service-key")
		event.AuditInfo([]string{clientIp, "%s", "%s %s as %s", status.Granted}, s.RefID, perms.First(), string(resource), s.GetClientRole().String())
		return s
	}

	// Find active session to perform authorization check or deny if no session was found.
	if s = Session(clientIp, authToken); s == nil {
		if s = authAnyJWT(c, clientIp, authToken, resource, perms); s != nil {
			event.AuditInfo([]string{clientIp, "session %s", "%s %s as %s", status.Granted}, s.RefID, perms.First(), string(resource), s.GetClientRole().String())
			return s
		}
		event.AuditWarn([]string{clientIp, "%s %s without authentication", status.Denied}, perms.String(), string(resource))
		return entity.SessionStatusUnauthorized()
	}

	// Set client IP.
	s.SetClientIP(clientIp)

	// If the request is from a client application, check its authorization based
	// on the allowed scope, the ACL, and the user account it belongs to (if any).
	if s.IsClient() {
		// Check the resource and required permissions against the session scope.
		if s.InsufficientScope(resource, perms) {
			event.AuditErr([]string{clientIp, "client %s", "session %s", "access %s", status.Error(authn.ErrInsufficientScope)}, clean.Log(s.GetClientInfo()), s.RefID, string(resource))
			return entity.SessionStatusForbidden()
		}

		// Check request authorization against client application ACL rules.
		if acl.Rules.DenyAll(resource, s.GetClientRole(), perms) {
			event.AuditErr([]string{clientIp, "client %s", "session %s", "%s %s", status.Denied}, clean.Log(s.GetClientInfo()), s.RefID, perms.String(), string(resource))
			return entity.SessionStatusForbidden()
		}

		// Also check the request authorization against the user's ACL rules?
		if s.NoUser() {
			// Allow access based on the ACL defaults for client applications.
			event.AuditInfo([]string{clientIp, "client %s", "session %s", "%s %s", status.Granted}, clean.Log(s.GetClientInfo()), s.RefID, perms.String(), string(resource))
		} else if u := s.GetUser(); !u.IsDisabled() && !u.IsUnknown() && u.IsRegistered() {
			if acl.Rules.DenyAll(resource, u.AclRole(), perms) {
				event.AuditErr([]string{clientIp, "client %s", "session %s", "%s %s as %s", status.Denied}, clean.Log(s.GetClientInfo()), s.RefID, perms.String(), string(resource), u.String())
				return entity.SessionStatusForbidden()
			}

			// Allow access based on the user role.
			event.AuditInfo([]string{clientIp, "client %s", "session %s", "%s %s as %s", status.Granted}, clean.Log(s.GetClientInfo()), s.RefID, perms.String(), string(resource), u.String())
		} else {
			// Deny access if it is not a regular user account or the account has been disabled.
			event.AuditErr([]string{clientIp, "client %s", "session %s", "%s %s as unauthorized user", status.Denied}, clean.Log(s.GetClientInfo()), s.RefID, perms.String(), string(resource))
			return entity.SessionStatusForbidden()
		}

		return s
	}

	// Perform a regular ACL authorization check based on the user role.
	u := s.GetUser()

	// Reject requests from unknown or disabled users.
	if u.IsUnknown() || u.IsDisabled() {
		event.AuditWarn([]string{clientIp, "session %s", "%s %s as unauthorized user", status.Denied}, s.RefID, perms.String(), string(resource))
		return entity.SessionStatusUnauthorized()
	}

	// Perform session scope check.
	if s.HasScope() {
		if s.InsufficientScope(resource, perms) {
			event.AuditErr([]string{clientIp, "session %s", "access %s", status.Error(authn.ErrInsufficientScope)}, s.RefID, string(resource))
			return entity.SessionStatusForbidden()
		}
	}

	// Perform ACL authorization check.
	if acl.Rules.DenyAll(resource, u.AclRole(), perms) {
		event.AuditErr([]string{clientIp, "session %s", "%s %s as %s", status.Denied}, s.RefID, perms.String(), string(resource), u.AclRole().String())
		return entity.SessionStatusForbidden()
	}

	// Permit access if all checks pass.
	event.AuditInfo([]string{clientIp, "session %s", "%s %s as %s", status.Granted}, s.RefID, perms.String(), string(resource), u.AclRole().String())
	return s
}

// AuthToken returns the client authentication token from the request context if one was found,
// or an empty string if no supported request header value was provided.
func AuthToken(c *gin.Context) string {
	return header.AuthToken(c)
}
