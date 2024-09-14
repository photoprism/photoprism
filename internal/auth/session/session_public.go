package session

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// PublicAuthToken is a static authentication token used in public mode.
var PublicAuthToken = "234200000000000000000000000000000000000000000000"

// PublicID is the SHA256 hash of the PublicAuthToken:
// a9b8ff820bf40ab451910f8bbfe401b2432446693aa539538fbd2399560a722f
var PublicID = rnd.SessionID(PublicAuthToken)

// public references the existing public mode session entity.
var public *entity.Session

// Public returns a client session for use in public mode.
func (s *Session) Public() *entity.Session {
	if public == nil {
		// Do nothing.
	} else if !public.Expired() {
		return public
	}

	public = entity.NewSession(0, 0)
	public.SetAuthToken(PublicAuthToken)
	public.AuthMethod = "public"
	public.SetUser(&entity.Admin)
	public.CacheDuration(-1)

	return public
}
