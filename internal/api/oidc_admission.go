package api

import (
	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
)

// OIDCSessionEligible reports whether a session may authorize a login at the provider endpoints.
// Only an interactive user session qualifies, and only while its account and its own scope still
// permit the cluster access the provider would delegate.
func OIDCSessionEligible(sess *entity.Session) bool {
	switch {
	case sess == nil || sess.Invalid() || sess.Expired() || sess.NoUser():
		return false
	case sess.IsClient():
		// Covers every client provider: OAuth clients, app passwords and access tokens.
		return false
	case sess.HasScope() && sess.InsufficientScope(acl.ResourceCluster, acl.Permissions{acl.ActionView}):
		return false
	}

	return sess.GetUser().CanLogIn()
}
