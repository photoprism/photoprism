package event

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/pkg/dummy"
)

// init initializes the event package.
func init() {
	// Event hooks for the default logger.
	hooks := logrus.LevelHooks{}
	hooks.Add(NewHook(SharedHub()))

	// Log is the global default logger.
	Log = &logrus.Logger{
		Out:          os.Stderr,
		Formatter:    TextFormatter,
		Hooks:        hooks,
		Level:        logrus.DebugLevel,
		ExitFunc:     os.Exit,
		ReportCaller: false,
	}

	// Create dummy audit logger.
	AuditLog = dummy.NewLogger()
}
