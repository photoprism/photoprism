package workers

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log

	// Remove temporary SQLite files before running the tests.
	fs.PurgeTestDbFiles(".", false)

	c := config.TestConfig()

	get.SetConfig(c)
	photoprism.SetConfig(c)

	// Run unit tests.
	code := m.Run()

	if err := c.CloseDb(); err != nil {
		log.Errorf("close db: %v", err)
	}

	// Remove temporary SQLite files after running the tests.
	fs.PurgeTestDbFiles(".", false)

	os.Exit(code)
}
