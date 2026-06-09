package event

import (
	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/auth/acl"
)

// SystemLog optionally records internal system events (background jobs, maintenance tasks).
var SystemLog Logger

// SystemPrefix prefixes single-segment messages sent to SystemLog.
var SystemPrefix = "system: "

// systemPrefixSep separates a multi-segment event's leading category from the
// rest of the message, e.g. the ": " in "config: database › connect".
const systemPrefixSep = ": "

// System writes a system-level log entry and publishes it to the hub.
//
// With more than one segment, the first segment becomes the log prefix
// (e.g. "config: database › connect"); a single segment keeps the generic
// SystemPrefix ("system: something happened"). The leading segment is treated
// as a plain category label and must not contain format verbs, since args are
// applied to the remaining segments.
func System(level logrus.Level, ev []string, args ...any) {
	if len(ev) == 0 {
		return
	}

	// Render the complete message (all segments joined) for the event hub so
	// the frontend log viewer keeps the leading category.
	message := Format(ev, args...)

	if SystemLog != nil {
		if len(ev) > 1 {
			SystemLog.Log(level, ev[0]+systemPrefixSep+Format(ev[1:], args...))
		} else {
			SystemLog.Log(level, SystemPrefix+message)
		}
	}

	Publish(
		string(acl.ChannelSystem)+".log."+level.String(),
		Data{
			"time":    TimeStamp(),
			"level":   level.String(),
			"message": message,
		},
	)
}

// SystemDebug records a system debug message.
func SystemDebug(ev []string, args ...any) {
	System(logrus.DebugLevel, ev, args...)
}

// SystemInfo records a system info message.
func SystemInfo(ev []string, args ...any) {
	System(logrus.InfoLevel, ev, args...)
}

// SystemWarn records a system warning.
func SystemWarn(ev []string, args ...any) {
	System(logrus.WarnLevel, ev, args...)
}

// SystemError records a system error message.
func SystemError(ev []string, args ...any) {
	System(logrus.ErrorLevel, ev, args...)
}
