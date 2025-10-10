package event

import (
	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/auth/acl"
)

// AuditLog optionally logs security events.
var AuditLog Logger

// AuditPrefix is prepended to audit log messages.
var AuditPrefix = "audit: "

// Audit optionally reports security-relevant events.
func Audit(level logrus.Level, ev []string, args ...interface{}) {
	// Skip if empty.
	if len(ev) == 0 {
		return
	}

	// Format log message.
	message := Format(ev, args...)

	// Show log message if AuditLog is specified.
	if AuditLog != nil {
		AuditLog.Log(level, AuditPrefix+message)
	}

	// Publish event if log level is info or higher.
	if level <= logrus.InfoLevel {
		Publish(
			string(acl.ChannelAudit)+".log."+level.String(),
			Data{
				"time":    TimeStamp(),
				"level":   level.String(),
				"message": message,
			},
		)
	}
}

// AuditErr records an audit entry at error level.
func AuditErr(ev []string, args ...interface{}) {
	Audit(logrus.ErrorLevel, ev, args...)
}

// AuditWarn records an audit entry at warning level.
func AuditWarn(ev []string, args ...interface{}) {
	Audit(logrus.WarnLevel, ev, args...)
}

// AuditInfo records an audit entry at info level.
func AuditInfo(ev []string, args ...interface{}) {
	Audit(logrus.InfoLevel, ev, args...)
}

// AuditDebug records an audit entry at debug level.
func AuditDebug(ev []string, args ...interface{}) {
	Audit(logrus.DebugLevel, ev, args...)
}
