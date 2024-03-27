package api

import (
	"github.com/photoprism/photoprism/internal/event"
)

var log = event.Log

// logErr logs an error if err is not nil.
func logErr(prefix string, err error) {
	if err != nil {
		log.Errorf("%s: %s", prefix, err.Error())
	}
}

// logWarn logs a warning if err is not nil.
func logWarn(prefix string, err error) {
	if err != nil {
		log.Warnf("%s: %s", prefix, err.Error())
	}
}
