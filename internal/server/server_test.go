package server

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/internal/testextras"
)

func TestMain(m *testing.M) {
	// Init test logger.
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log

	caller := "internal/server/server_test.go/TestMain"
	dbc, err := testextras.AcquireDBMutex(log, caller)
	if err != nil {
		log.Error("FAIL")
		os.Exit(1)
	}
	defer testextras.UnlockDBMutex(dbc.Db())

	// Init test config.
	c := config.TestConfig()
	get.SetConfig(c)
	defer c.CloseDb()

	// Increase login rate limit for testing.
	limiter.Login = limiter.NewLimit(1, 10000)

	// Run unit tests.
	code := m.Run()

	testextras.ReleaseDBMutex(dbc.Db(), log, caller, code)

	os.Exit(code)
}
