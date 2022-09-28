package session

import (
	"time"

	"github.com/photoprism/photoprism/internal/config"
)

// New returns a new session store with an optional cachePath.
func New(expiresAfter time.Duration, conf *config.Config) *Session {
	return &Session{expiresAfter: expiresAfter, conf: conf}
}
