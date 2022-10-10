package sev

import (
	"fmt"
)

const (
	Emergency Level = iota
	Alert
	Critical
	Error
	Warning
	Notice
	Info
	Debug
)

var Levels = []Level{
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
func (level *Level) UnmarshalText(text []byte) error {
	l, err := Parse(string(text))
	if err != nil {
		return err
	}

	*level = l

	return nil
}

func (level Level) MarshalText() ([]byte, error) {
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

func (level Level) Status() string {
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
