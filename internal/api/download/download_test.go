package download

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/fs"
)

// TestMain executes runTestMain returning it's results.  It is done this way so that defer can be used to cleanup.
func TestMain(m *testing.M) {
	os.Exit(runTestMain(m))
}

func runTestMain(m *testing.M) int {
	// Init test logger.
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log
	AllowedPaths = append(AllowedPaths, fs.Abs("./testdata"))

	// Run unit tests.
	return m.Run()
}
