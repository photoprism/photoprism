package download

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestMain(m *testing.M) {
	// Init test logger.
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log
	AllowedPaths = append(AllowedPaths, fs.Abs("./testdata"))

	// Run unit tests.
	code := m.Run()

	os.Exit(code)
}
