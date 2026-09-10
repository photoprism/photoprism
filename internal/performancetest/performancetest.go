package performancetest

import (
	"time"

	"github.com/photoprism/photoprism/internal/event"
)

var log = event.Log

// Log logs the error if any and keeps quiet otherwise.
func Log(model, action string, err error) {
	if err != nil {
		log.Errorf("%s: %s (%s)", model, err, action)
	}
}

// UTC returns the current Coordinated Universal Time (UTC).
func UTC() time.Time {
	return time.Now().UTC()
}
