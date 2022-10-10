package event

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// AuditLog optionally logs security events.
var AuditLog Logger
var AuditPrefix = "audit: "
var AuditMessageSep = " › "

// Format formats an audit log event.
func Format(ev []string, args ...interface{}) string {
	return fmt.Sprintf(strings.Join(ev, AuditMessageSep), args...)
}

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
			"audit."+level.String(),
			Data{
				"time":    TimeStamp(),
				"level":   level.String(),
				"message": message,
			},
		)
	}
}

func AuditErr(ev []string, args ...interface{}) {
	Audit(logrus.ErrorLevel, ev, args...)
}

func AuditWarn(ev []string, args ...interface{}) {
	Audit(logrus.WarnLevel, ev, args...)
}

func AuditInfo(ev []string, args ...interface{}) {
	Audit(logrus.InfoLevel, ev, args...)
}

func AuditDebug(ev []string, args ...interface{}) {
	Audit(logrus.DebugLevel, ev, args...)
}
