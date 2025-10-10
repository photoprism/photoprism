package config

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/event"
)

// log points to the global logger.
var log = event.Log

// SetLogLevel sets the application log level.
func SetLogLevel(level logrus.Level) {
	SetTensorFlowLogLevel(level)
	log.SetLevel(level)
	if event.SystemLog != nil {
		event.SystemLog.SetLevel(level)
	}
}

// SetTensorFlowLogLevel sets the TensorFlow log level.
func SetTensorFlowLogLevel(level logrus.Level) {
	switch level {
	case logrus.TraceLevel:
		_ = os.Setenv("TF_CPP_MIN_LOG_LEVEL", "0")
	case logrus.DebugLevel:
		_ = os.Setenv("TF_CPP_MIN_LOG_LEVEL", "1")
	case logrus.InfoLevel:
		_ = os.Setenv("TF_CPP_MIN_LOG_LEVEL", "2")
	case logrus.WarnLevel:
		_ = os.Setenv("TF_CPP_MIN_LOG_LEVEL", "3")
	case logrus.ErrorLevel:
		_ = os.Setenv("TF_CPP_MIN_LOG_LEVEL", "4")
	case logrus.FatalLevel, logrus.PanicLevel:
		_ = os.Setenv("TF_CPP_MIN_LOG_LEVEL", "5")
	}
}
