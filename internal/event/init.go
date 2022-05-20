package event

import (
	"os"

	"github.com/sirupsen/logrus"
)

func init() {
	hooks := logrus.LevelHooks{}
	hooks.Add(NewHook(SharedHub()))

	Log = &logrus.Logger{
		Out:          os.Stderr,
		Formatter:    &logrus.TextFormatter{},
		Hooks:        hooks,
		Level:        logrus.DebugLevel,
		ExitFunc:     os.Exit,
		ReportCaller: false,
	}
}
