package clean

import "strings"

// Error sanitizes an error message so that it can be safely logged or displayed.
func Error(err error) string {
	if err == nil {
		return "no error"
	} else if s := strings.TrimSpace(err.Error()); s == "" {
		return "unknown error"
	} else {
		return Log(s)
	}
}
