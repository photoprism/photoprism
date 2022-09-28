package event

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// AuditLog optionally logs security events.
var AuditLog Logger
var AuditPrefix = "audit: "

// Format formats an audit log event.
func Format(ev []string, args ...interface{}) string {
	return fmt.Sprintf(strings.Join(ev, " › "), args...)
}

// Audit optionally reports security-relevant events.
func Audit(level logrus.Level, ev []string, args ...interface{}) {
	if AuditLog != nil && len(ev) > 0 {
		AuditLog.Log(level, AuditPrefix+Format(ev, args...))
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
