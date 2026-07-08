package event

import "runtime/debug"

// LogPanic reports a recovered panic value together with the current stack
// trace so that crashes which would otherwise terminate the process without any
// output can be diagnosed. It does not stop the process; the caller decides
// whether to exit.
//
// It deliberately reports through SystemError (the system log) rather than the
// default Log: the default logger is persisted to the database via the event
// hub (the errors and audit_logs tables), so reporting a panic through it could
// trigger a follow-up error or panic when the database is itself the cause of
// the crash or is unavailable. The system log writes to the console only.
func LogPanic(r any) {
	if r == nil {
		return
	}

	// The leading "panic" segment becomes the log prefix, so this reads as
	// "panic: <value> › <stack>".
	SystemError([]string{"panic", "%v", "%s"}, r, debug.Stack())
}

// Recover recovers from a panic in the calling goroutine, if any, and logs it
// with a stack trace via LogPanic. Use it as a deferred call so a crash in a
// background task is reported instead of silently terminating the process:
// defer event.Recover().
func Recover() {
	if r := recover(); r != nil {
		LogPanic(r)
	}
}

// Safe runs fn and recovers from any panic it raises, logging it via LogPanic.
// It lets long-running loops continue after a single iteration panics rather
// than taking down the whole process.
func Safe(fn func()) {
	defer Recover()
	fn()
}
