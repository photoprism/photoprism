package level

import (
	"fmt"
)

// Severity levels.
const (
	Emergency Severity = iota
	Alert
	Critical
	Error
	Warning
	Notice
	Info
	Debug
)

// Levels contains the valid severity levels.
var Levels = []Severity{
	Emergency,
	Alert,
	Critical,
	Error,
	Warning,
	Notice,
	Info,
	Debug,
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (level *Severity) UnmarshalText(text []byte) error {
	l, err := Parse(string(text))
	if err != nil {
		return err
	}

	*level = l

	return nil
}

// MarshalText implements encoding.TextMarshaler.
func (level Severity) MarshalText() ([]byte, error) {
	switch level {
	case Debug:
		return []byte("debug"), nil
	case Info:
		return []byte("info"), nil
	case Notice:
		return []byte("notice"), nil
	case Warning:
		return []byte("warning"), nil
	case Error:
		return []byte("error"), nil
	case Critical:
		return []byte("critical"), nil
	case Alert:
		return []byte("alert"), nil
	case Emergency:
		return []byte("emergency"), nil
	}

	return nil, fmt.Errorf("not a valid severity level %d", level)
}

// Status returns the severity level as an info string for reports.
func (level Severity) Status() string {
	switch level {
	case Warning:
		return "warning"
	case Error:
		return "error"
	case Critical:
		return "critical"
	case Alert:
		return "alert"
	case Emergency:
		return "emergency"
	default:
		return "OK"
	}
}
