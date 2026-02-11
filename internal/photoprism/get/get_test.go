package get

import (
	"os"
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/testextras"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestMain(m *testing.M) {

	caller := "internal/photoprism/get/get_test.go/TestMain"
	dbc, dbn, err := testextras.AcquireDBMutex(log, caller)
	if err != nil {
		log.Error("FAIL")
		os.Exit(1)
	}
	defer testextras.UnlockDBMutex(dbc.Db())

	_, dsname := dsn.PhotoPrismTestToDriverDSN(dbn)
	dsn.SetDSNToEnv(dsname)

	tempDir, err := os.MkdirTemp("", "internal-photoprism-get")
	if err != nil {
		panic(err)
	}
	c := config.NewMinimalTestConfigWithDb("test", tempDir)

	SetConfig(c)

	beforeTimestamp := time.Now().UTC()
	code := m.Run()
	code = testextras.ValidateDBErrors(c.Db(), log, beforeTimestamp, code)

	testextras.ReleaseDBMutex(dbc.Db(), log, caller, code)

	if err = c.CloseDb(); err != nil {
		log.Warnf("close db: %v", err)
	}

	if err = os.RemoveAll(tempDir); err != nil {
		log.Errorf("remove temp dir: %v", err)
	}

	// Remove temporary SQLite files after running the tests.
	fs.PurgeTestDbFiles(".", false)

	os.Exit(code)
}
