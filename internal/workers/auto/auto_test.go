package auto

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/config"
)

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)

	c := config.TestConfig()
	defer c.CloseDb()

	// Run unit tests.
	code := m.Run()

	// Close database connection.
	_ = c.CloseDb()

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
