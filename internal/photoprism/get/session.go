package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/auth/session"
)

var onceSession sync.Once

func initSession() {
	services.Session = session.New(Config())
}

func Session() *session.Session {
	onceSession.Do(initSession)

	return services.Session
}
