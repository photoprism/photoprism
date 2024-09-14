package session

import (
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
)

var stop = make(chan bool, 1)

// CleanupAction deletes sessions that have expired.
var CleanupAction = func() {
	if n := entity.DeleteExpiredSessions(); n > 0 {
		event.AuditInfo([]string{"deleted %s"}, english.Plural(n, "expired session", "expired sessions"))
	} else {
		event.AuditDebug([]string{"found no expired sessions"})
	}
}

// Cleanup starts a background worker that periodically deletes expired sessions.
func Cleanup(interval time.Duration) {
	// Immediately delete sessions that have already expired.
	CleanupAction()

	// Periodically delete expired sessions based on the specified interval.
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-stop:
				ticker.Stop()
				return
			case <-ticker.C:
				CleanupAction()
			}
		}
	}()
}

// Shutdown shuts down the session watcher.
func Shutdown() {
	stop <- true
}
