package service

import (
	"sync"

	"github.com/photoprism/photoprism/internal/session"
)

var onceSession sync.Once

func initSession() {
	// Sessions are valid for 7 days by default.
	services.Session = session.New(session.ExpiresAfter, Config())
}

func Session() *session.Session {
	onceSession.Do(initSession)

	return services.Session
}
