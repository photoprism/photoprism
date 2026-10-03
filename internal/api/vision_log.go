package api

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
)

// logVisionErr logs a failed vision request with a fixed message and writes its error text to the
// system log with URI userinfo and queries redacted.
func logVisionErr(action string, err error) {
	if err == nil {
		return
	}

	text := clipTokens(err.Error(), 4*clean.LengthLimit)
	text = clipTokens(clean.UriQueriesRedacted(clean.UriRedactedText(text)), clean.LengthLimit)

	log.Errorf("vision: %s request failed (details in system log)", action)
	event.SystemError([]string{"vision", "%s request failed", "%s"}, action, clean.ErrorFull(errors.New(text)))
}

// clipTokens shortens s to at most n bytes, ending at the last URI delimiter (clean.UriDelimiter).
func clipTokens(s string, n int) string {
	if len(s) <= n {
		return s
	}

	if next, _ := utf8.DecodeRuneInString(s[n:]); clean.UriDelimiter(next) {
		return s[:n]
	}

	s = s[:n]

	if i := strings.LastIndexFunc(s, clean.UriDelimiter); i >= 0 {
		return s[:i]
	}

	return ""
}
