package auto

import (
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/testextras"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)

	caller := "internal/workers/auto/auto_test.go/TestMain"
	dbc, dbn, err := testextras.AcquireDBMutex(log, caller)
	if err != nil {
		log.Error("FAIL")
		os.Exit(1)
	}
	defer testextras.UnlockDBMutex(dbc.Db())

	_, dsname := dsn.PhotoPrismTestToDriverDSN(dbn)
	dsn.SetDSNToEnv(dsname)

	c := config.TestConfig()

	// Run unit tests.
	beforeTimestamp := time.Now().UTC()
	code := m.Run()
	code = testextras.ValidateDBErrors(c.Db(), log, beforeTimestamp, code)
	testextras.ReleaseDBMutex(dbc.Db(), log, caller, code)
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
