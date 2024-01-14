package api

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Session finds the client session for the specified
// auth token, or returns nil if not found.
func Session(authToken string) *entity.Session {
	// Skip authentication when running in public mode.
	if get.Config().Public() {
		return get.Session().Public()
	} else if !rnd.IsAuthAny(authToken) {
		return nil
	}

	// Find the session based on the hashed auth
	// token used as id, or return nil otherwise.
	if s, err := get.Session().Get(rnd.SessionID(authToken)); err != nil {
		return nil
	} else {
		return s
	}
}
