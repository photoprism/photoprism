package session

import (
	"github.com/photoprism/photoprism/internal/entity"
)

var Public *entity.Session
var PublicID = "234200000000000000000000000000000000000000000000"

// Public returns a client session for use in public mode.
func (s *Session) Public() *entity.Session {
	if Public == nil {
		// Do nothing.
	} else if !Public.Expired() {
		return Public
	}

	Public = entity.NewSession(0, 0)
	Public.ID = PublicID
	Public.AuthMethod = "public"
	Public.SetUser(&entity.Admin)
	Public.CacheDuration(-1)

	return Public
}
