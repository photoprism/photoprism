package get

import (
	"os"
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/testextras"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestMain(m *testing.M) {

	caller := "internal/photoprism/get/get_test.go/TestMain"
	dbc, err := testextras.AcquireDBMutex(log, caller)
	if err != nil {
		log.Error("FAIL")
		os.Exit(1)
	}
	defer testextras.UnlockDBMutex(dbc.Db())

	tempDir, err := os.MkdirTemp("", "internal-photoprism-get")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)
	c := config.NewMinimalTestConfigWithDb("test", tempDir)

	SetConfig(c)
	defer c.CloseDb()

	beforeTimestamp := time.Now().UTC()
	code := m.Run()
	code = testextras.ValidateDBErrors(dbc.Db(), log, beforeTimestamp, code)

	testextras.ReleaseDBMutex(dbc.Db(), log, caller, code)

	// Remove temporary SQLite files after running the tests.
	fs.PurgeTestDbFiles(".", false)

	os.Exit(code)
}
