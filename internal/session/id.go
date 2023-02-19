package session

import (
	"github.com/photoprism/photoprism/pkg/rnd"
)

// ID returns a random 48-character session id string.
func ID() string {
	return rnd.SessionID()
}
