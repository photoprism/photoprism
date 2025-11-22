package auto

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)

	c := config.TestConfig()

	// Run unit tests.
	code := m.Run()

	if err := c.CloseDb(); err != nil {
		log.Errorf("close db: %v", err)
	}

	// Remove temporary SQLite files after running the tests.
	fs.PurgeTestDbFiles(".", false)

	os.Exit(code)
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
