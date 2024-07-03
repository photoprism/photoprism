package level

import "github.com/sirupsen/logrus"

// Logrus takes a logrus.Level and returns the corresponding Severity.
func Logrus(level logrus.Level) Severity {
	switch level {
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
