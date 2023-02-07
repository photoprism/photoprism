package session

import (
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
)

var stop = make(chan bool, 1)

// MonitorAction deletes expired sessions.
var MonitorAction = func() {
	if n := entity.DeleteExpiredSessions(); n > 0 {
		event.AuditInfo([]string{"deleted %s"}, english.Plural(n, "expired session", "expired sessions"))
	} else {
		event.AuditDebug([]string{"found no expired sessions"})
	}
}

// Monitor starts a background worker that periodically deletes expired sessions.
func Monitor(interval time.Duration) {
	ticker := time.NewTicker(interval)

	MonitorAction()

	go func() {
		for {
			select {
			case <-stop:
				ticker.Stop()
				return
			case <-ticker.C:
				MonitorAction()
			}
		}
	}()
}

// Shutdown shuts down the session watcher.
func Shutdown() {
	stop <- true
}
