package session

import (
	"time"

	"github.com/photoprism/photoprism/internal/config"
)

// MaxAge is the maximum duration after which a session expires.
var MaxAge = 168 * time.Hour * 24 * 7
var Timeout = 168 * time.Hour * 24 * 3

// New creates a new session store with default values.
func New(conf *config.Config) *Session {
	return &Session{MaxAge: MaxAge, Timeout: Timeout, conf: conf}
}
