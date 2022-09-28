package session

import (
	"github.com/photoprism/photoprism/internal/entity"
)

var Public *entity.Session
var PublicID = "234200000000000000000000000000000000000000000000"

// Public returns a client session for use in public mode.
func (s *Session) Public() *entity.Session {
	if Public != nil {
		return Public
	}

	Public = entity.NewSession(s.expiresAfter)
	Public.ID = PublicID
	Public.SetUser(&entity.Admin)
	Public.CacheDuration(-1)

	return Public
}
