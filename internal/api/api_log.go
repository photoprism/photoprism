package api

import (
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
)

var log = event.Log

// logErr logs an error if err is not nil.
func logErr(prefix string, err error) {
	if err != nil {
		log.Errorf("%s: %s", prefix, clean.Error(err))
	}
}

// systemErr reports an error on the console channel if err is not nil, for a failure that only an
// operator acts on. It renders the full cause, since that reader wants the location, and it is not
// written to the errors table the way a log entry is. The prefix stays a segment because System
// renders the first one outside Format.
func systemErr(prefix string, err error) {
	if err != nil {
		event.SystemError([]string{prefix, "%s"}, clean.ErrorFull(err))
	}
}

// logWarn logs a warning if err is not nil.
func logWarn(prefix string, err error) {
	if err != nil {
		log.Warnf("%s: %s", prefix, clean.Error(err))
	}
}
