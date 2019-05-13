package util

import (
	"time"

	log "github.com/sirupsen/logrus"
)

func ProfileTime(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Debugf("%s took %s", name, elapsed)
}
