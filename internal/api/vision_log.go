package api

import (
	"errors"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
)

// logVisionErr logs a failed vision request with a fixed message and writes its error text to the
// system log with URI userinfo and queries redacted.
func logVisionErr(action string, err error) {
	if err == nil {
		return
	}

	text := err.Error()

	if len(text) > 2*clean.LengthLimit {
		text = text[:2*clean.LengthLimit]
	}

	log.Errorf("vision: %s request failed (details in system log)", action)
	event.SystemError([]string{"vision", "%s request failed", "%s"}, action, clean.ErrorFull(errors.New(clean.UriQueriesRedacted(clean.UriRedactedText(text)))))
}
