package auto

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/testextras"
	"github.com/photoprism/photoprism/pkg/fs"
)

// TestMain executes runTestMain returning it's results.  It is done this way so that defer can be used to cleanup.
func TestMain(m *testing.M) {
	os.Exit(runTestMain(m))
}

func runTestMain(m *testing.M) (code int) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)

	// Remove temporary SQLite files before running the tests.
	fs.PurgeTestDbFiles(".", false)

	c := config.TestConfig()
	defer c.CleanupTestFolder()
	defer func() {
		if err := c.CloseDb(); err != nil {
			log.Warnf("close db: %v", err)
		}
		// Remove temporary SQLite files after running the tests.
		fs.PurgeTestDbFiles(".", false)
	}()

	// Run unit tests.
	return testextras.TestDbCleanup(m.Run())
}

func TestStart(t *testing.T) {
	conf := config.TestConfig()

	Start(conf)
	ShouldIndex()
	ShouldImport()

	if mustIndex(conf.AutoIndex()) {
		t.Error("mustIndex() must return false")
	}

	if mustImport(conf.AutoImport()) {
		t.Error("mustImport() must return false")
	}

	ResetImport()
	ResetIndex()

	Shutdown()
}
