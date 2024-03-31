package api

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Session finds the client session for the specified auth token, or returns nil if not found.
func Session(clientIp, authToken string) *entity.Session {
	// Skip authentication when running in public mode.
	if get.Config().Public() {
		return get.Session().Public()
	}

	// Fail if the auth token does not have a supported format.
	if !rnd.IsAuthAny(authToken) {
		return nil
	}

	// Check request rate limit.
	r := limiter.Auth.Request(clientIp)

	if r.Reject() {
		return nil
	}

	// Find the session based on the hashed auth token, or return nil otherwise.
	s, err := entity.FindSession(rnd.SessionID(authToken))

	if err != nil {
		return nil
	}

	// Return the reserved request rate limit tokens after successful authentication.
	r.Success()

	return s
}
