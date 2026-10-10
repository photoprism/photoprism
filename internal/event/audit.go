package event

import (
	"net"
	"strings"
	"sync/atomic"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/auth/acl"
)

// AuditLog optionally logs security events.
var AuditLog Logger

// AuditPrefix is prepended to audit log messages.
var AuditPrefix = "audit: "

// AuditRecorder persists an audit event when it is emitted.
type AuditRecorder func(data Data)

// auditRecorder holds the recorder that persists audit events synchronously, if any.
var auditRecorder atomic.Pointer[AuditRecorder]

// SetAuditRecorder sets the recorder that persists audit events when they are emitted, so that short-lived
// processes such as CLI commands record them before they exit, and returns the previous one. Passing nil
// leaves recording to the subscribers of the event hub.
func SetAuditRecorder(r AuditRecorder) (prev AuditRecorder) {
	var old *AuditRecorder

	if r == nil {
		old = auditRecorder.Swap(nil)
	} else {
		old = auditRecorder.Swap(&r)
	}

	if old == nil {
		return nil
	}

	return *old
}

// AuditRecordedSync reports whether audit events are persisted when they are emitted, so that event hub
// subscribers must not persist them again.
func AuditRecordedSync() bool {
	return auditRecorder.Load() != nil
}

// Audit optionally reports security-relevant events.
func Audit(level logrus.Level, ev []string, args ...any) {
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

	// Record and publish event if log level is info or higher.
	if level <= logrus.InfoLevel {
		data := Data{
			"time":    TimeStamp(),
			"level":   level.String(),
			"ip":      AuditIP(ev),
			"message": message,
		}

		if r := auditRecorder.Load(); r != nil {
			(*r)(data)
		}

		Publish(string(acl.ChannelAudit)+".log."+level.String(), data)
	}
}

// AuditIP returns the client address of an audit event, which the Who-What-Outcome convention puts
// in its first segment. A segment that is not an address yields an empty string, so an event with no
// peer of its own reports none.
func AuditIP(ev []string) string {
	if len(ev) == 0 {
		return ""
	}

	if ip := net.ParseIP(strings.TrimSpace(ev[0])); ip != nil {
		return ip.String()
	}

	return ""
}

// AuditTrace records an audit entry at trace level.
func AuditTrace(ev []string, args ...any) {
	Audit(logrus.TraceLevel, ev, args...)
}

// AuditDebug records an audit entry at debug level.
func AuditDebug(ev []string, args ...any) {
	Audit(logrus.DebugLevel, ev, args...)
}

// AuditInfo records an audit entry at info level.
func AuditInfo(ev []string, args ...any) {
	Audit(logrus.InfoLevel, ev, args...)
}

// AuditWarn records an audit entry at warning level.
func AuditWarn(ev []string, args ...any) {
	Audit(logrus.WarnLevel, ev, args...)
}

// AuditErr records an audit entry at error level.
func AuditErr(ev []string, args ...any) {
	Audit(logrus.ErrorLevel, ev, args...)
}
