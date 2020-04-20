package service

import (
	"sync"
	"time"

	"github.com/photoprism/photoprism/internal/session"
)

var onceSession sync.Once

func initSession() {
	// keep sessions for 7 days by default
	services.Session = session.New(168*time.Hour, Config().CachePath())
}

func Session() *session.Session {
	onceSession.Do(initSession)

	return services.Session
}
