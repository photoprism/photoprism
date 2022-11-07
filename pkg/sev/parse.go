package sev

import (
	"fmt"
	"strings"
)

// Parse takes a string level and returns the severity constant.
func Parse(lvl string) (Level, error) {
	switch strings.ToLower(lvl) {
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

	var l Level
	return l, fmt.Errorf("not a valid Level: %q", lvl)
}
