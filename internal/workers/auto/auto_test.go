package auto

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/testextras"
)

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)

	caller := "internal/workers/auto/auto_test.go/TestMain"
	dbc, err := testextras.AcquireDBMutex(log, caller)
	if err != nil {
		log.Error("FAIL")
		os.Exit(1)
	}
	defer testextras.UnlockDBMutex(dbc.Db())

	c := config.TestConfig()
	defer c.CloseDb()

	// Run unit tests.
	code := m.Run()
	testextras.ReleaseDBMutex(dbc.Db(), log, caller, code)
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
