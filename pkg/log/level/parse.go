package level

import (
	"fmt"
	"strings"
)

// Parse takes a string and returns the corresponding severity, if any.
func Parse(level string) (Severity, error) {
	switch strings.ToLower(level) {
	case "emergency", "emerg", "panic":
		return Emergency, nil
	case "fatal", "alert":
		return Alert, nil
	case "critical", "crit":
		return Critical, nil
	case "error", "err":
		return Error, nil
	case "warn", "warning":
		return Warning, nil
	case "notice", "note":
		return Notice, nil
	case "info", "informational", "ok":
		return Info, nil
	case "debug":
		return Debug, nil
	}

	var l Severity
	return l, fmt.Errorf("not a valid severity level: %q", level)
}
