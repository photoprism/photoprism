package event

import (
	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/auth/acl"
)

// SystemLog optionally records internal system events (background jobs, maintenance tasks).
var SystemLog Logger

// SystemPrefix prefixes messages sent to SystemLog.
var SystemPrefix = "system: "

// System writes a system-level log entry and publishes it to the hub.
func System(level logrus.Level, ev []string, args ...interface{}) {
	if len(ev) == 0 {
		return
	}

	message := Format(ev, args...)

	if SystemLog != nil {
		SystemLog.Log(level, SystemPrefix+message)
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

// SystemWarn records a system warning.
func SystemWarn(ev []string, args ...interface{}) {
	System(logrus.WarnLevel, ev, args...)
}

// SystemInfo records a system info message.
func SystemInfo(ev []string, args ...interface{}) {
	System(logrus.InfoLevel, ev, args...)
}

// SystemDebug records a system debug message.
func SystemDebug(ev []string, args ...interface{}) {
	System(logrus.DebugLevel, ev, args...)
}

// SystemError records a system error message.
func SystemError(ev []string, args ...interface{}) {
	System(logrus.ErrorLevel, ev, args...)
}
