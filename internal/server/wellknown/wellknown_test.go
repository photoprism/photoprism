package wellknown

import (
	"os"
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/testextras"
	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	// Init test logger.
	log := logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log

	caller := "internal/server/wellknwon/wellknown_test.go/TestMain"
	dbc, err := testextras.AcquireDBMutex(log, caller)
	if err != nil {
		log.Error("FAIL")
		os.Exit(1)
	}
	defer testextras.UnlockDBMutex(dbc.Db())

	// Run unit tests.
	beforeTimestamp := time.Now().UTC()
	code := m.Run()
	code = testextras.ValidateDBErrors(dbc.Db(), log, beforeTimestamp, code)

	testextras.ReleaseDBMutex(dbc.Db(), log, caller, code)

	os.Exit(code)
}
