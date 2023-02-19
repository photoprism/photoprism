package sev

import "github.com/sirupsen/logrus"

// LogLevel takes a logrus log level and returns the severity.
func LogLevel(lvl logrus.Level) Level {
	switch lvl {
	case logrus.PanicLevel:
		return Alert
	case logrus.FatalLevel:
		return Critical
	case logrus.ErrorLevel:
		return Error
	case logrus.WarnLevel:
		return Warning
	case logrus.InfoLevel:
		return Info
	default:
		return Debug
	}
}
