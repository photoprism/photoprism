package api

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Session finds the client session for the specified auth token, or returns nil if not found.
func Session(clientIp, authToken string) (sess *entity.Session) {
	// Skip authentication and return the default session when public mode is enabled.
	if get.Config().Public() {
		return get.Session().Public()
	}

	// Check auth token format and return nil if it is invalid.
	if !rnd.IsAuthAny(authToken) {
		return nil
	}

	// Check failure rate limit and return nil if it has been exceeded.
	if limiter.Auth.Reject(clientIp) {
		return nil
	}

	// Try to find an active session based on the hashed auth token.
	sess, err := entity.FindSession(rnd.SessionID(authToken))

	// Count error towards failure rate limit and return nil.
	if err != nil {
		limiter.Auth.Reserve(clientIp)
		return nil
	}

	// Return session.
	return sess
}
