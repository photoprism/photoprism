package report

import (
	"time"

	"github.com/photoprism/photoprism/pkg/txt"
)

// DateTime formats a time pointer as a human-readable datetime string.
func DateTime(t *time.Time) string {
	return txt.DateTime(t)
}

// Date formats a time pointer as a date without a time of day, for values that carry none - a
// rendered midnight would invent precision. Empty for a nil or zero time, which is how "not set"
// is stored.
func Date(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}

	return t.Format("2006-01-02")
}

// UnixTime formats a unix time as a human-readable datetime string.
func UnixTime(t int64) string {
	return txt.UnixTime(t)
}
