package get

import (
	"os"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/testextras"
)

func TestMain(m *testing.M) {

	caller := "internal/photoprism/get/get_test.go/TestMain"
	dbc, err := testextras.AcquireDBMutex(log, caller)
	if err != nil {
		log.Error("FAIL")
		os.Exit(1)
	}
	defer testextras.UnlockDBMutex(dbc.Db())

	c := config.NewTestConfig("service")
	SetConfig(c)
	defer c.CloseDb()

	code := m.Run()

	testextras.ReleaseDBMutex(dbc.Db(), log, caller, code)

	os.Exit(code)
}
