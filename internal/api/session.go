package api

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Session finds the client session for the specified auth token, or returns nil if not found.
func Session(clientIp, authToken string) *entity.Session {
	sess, _ := LookupSession(clientIp, authToken)
	return sess
}

// LookupSession finds the client session for the specified auth token, or returns nil and an error that
// names the reason: authn.ErrTokenRequired, authn.ErrInvalidToken, or authn.ErrRateLimitExceeded.
// A failed database query also returns authn.ErrInvalidToken, without counting against the limit.
func LookupSession(clientIp, authToken string) (*entity.Session, error) {
	return lookupSession(clientIp, authToken, "")
}

// lookupSession finds the session for the auth token like LookupSession, and also accepts a session that was
// issued for the specified audience if it is not empty.
func lookupSession(clientIp, authToken, audience string) (*entity.Session, error) {
	// Skip authentication and return the default session when public mode is enabled.
	if get.Config().Public() {
		return get.Session().Public(), nil
	}

	// Check auth token format and return nil if it is invalid.
	if authToken == "" {
		return nil, authn.ErrTokenRequired
	} else if !rnd.IsAuthAny(authToken) {
		return nil, authn.ErrInvalidToken
	}

	// Check failure rate limit and return nil if it has been exceeded.
	if limiter.Auth.Reject(clientIp) {
		return nil, authn.ErrRateLimitExceeded
	}

	// Try to find an active session based on the hashed auth token.
	var sess *entity.Session
	var err error

	if audience == "" {
		sess, err = entity.FindSession(rnd.SessionID(authToken))
	} else {
		sess, err = entity.FindAudienceSession(rnd.SessionID(authToken), audience)
	}

	// Count a token without an active session towards the failure rate limit and return nil.
	if err != nil {
		if entity.SessionNotFound(err) {
			limiter.Auth.Reserve(clientIp)
		}

		return nil, authn.ErrInvalidToken
	}

	// Return session.
	return sess, nil
}
