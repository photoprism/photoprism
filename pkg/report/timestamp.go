package report

import (
	"time"

	"github.com/photoprism/photoprism/pkg/txt"
)

// DateTime formats a time pointer as a human-readable datetime string.
func DateTime(t *time.Time) string {
	return txt.DateTime(t)
}

// UnixTime formats a unix time as a human-readable datetime string.
func UnixTime(t int64) string {
	return txt.UnixTime(t)
}
