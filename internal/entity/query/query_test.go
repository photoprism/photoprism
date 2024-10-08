package query

import (
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/testextras"
)

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)

	caller := "internal/entity/query/query_test.go/TestMain"
	dbc, err := testextras.AcquireDBMutex(log, caller)
	if err != nil {
		log.Error("FAIL")
		os.Exit(1)
	}
	defer testextras.UnlockDBMutex(dbc.Db())

	db := entity.InitTestDb(
		os.Getenv("PHOTOPRISM_TEST_DRIVER"),
		os.Getenv("PHOTOPRISM_TEST_DSN"))

	defer db.Close()

	beforeTimestamp := time.Now().UTC()
	code := m.Run()
	code = testextras.ValidateDBErrors(dbc.Db(), log, beforeTimestamp, code)

	testextras.ReleaseDBMutex(dbc.Db(), log, caller, code)

	os.Exit(code)
}

func TestDbDialect(t *testing.T) {
	t.Run("SQLite", func(t *testing.T) {
		if DbDialect() != SQLite3 {
			t.SkipNow()
		}
		assert.Equal(t, "sqlite", DbDialect())
	})

	t.Run("MariaDB", func(t *testing.T) {
		if DbDialect() != MySQL {
			t.SkipNow()
		}
		assert.Equal(t, MySQL, DbDialect())
	})
}

func TestBatchSize(t *testing.T) {
	t.Run("SQLite", func(t *testing.T) {
		if DbDialect() != SQLite3 {
			t.SkipNow()
		}
		assert.Equal(t, 333, BatchSize())
	})
}
