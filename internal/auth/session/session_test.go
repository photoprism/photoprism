package session

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log

	c := config.TestConfig()

	code := m.Run()

	// Remove temporary SQLite files after running the tests.
	if err := c.CloseDb(); err != nil {
		log.Errorf("close db: %v", err)
	}

	fs.PurgeTestDbFiles(".", false)

	os.Exit(code)
}
