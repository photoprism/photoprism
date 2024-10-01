package entity

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/testextras"
)

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log

	caller := "internal/entity/entity_test.go/TestMain"
	dbc, err := testextras.AcquireDBMutex(log, caller)
	if err != nil {
		log.Error("FAIL")
		os.Exit(1)
	}
	defer testextras.UnlockDBMutex(dbc.Db())

	db := InitTestDb(
		os.Getenv("PHOTOPRISM_TEST_DRIVER"),
		os.Getenv("PHOTOPRISM_TEST_DSN"))

	defer db.Close()

	code := m.Run()

	testextras.ReleaseDBMutex(dbc.Db(), log, caller, code)

	os.Exit(code)
}

func TestTypeString(t *testing.T) {
	assert.Equal(t, "unknown", TypeString(""))
	assert.Equal(t, "foo", TypeString("foo"))
}
