package avatar

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/testextras"
)

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)

	caller := "internal/thumb/avatar/avatar_test.go/TestMain"
	dbc, err := testextras.AcquireDBMutex(log, caller)
	if err != nil {
		log.Error("FAIL")
		os.Exit(1)
	}
	defer testextras.UnlockDBMutex(dbc.Db())

	c := config.NewTestConfig("avatar")
	get.SetConfig(c)
	photoprism.SetConfig(c)
	defer c.CloseDb()

	code := m.Run()

	testextras.ReleaseDBMutex(dbc.Db(), log, caller, code)

	os.Exit(code)
}
