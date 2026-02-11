package query

import (
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/testextras"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
)

// staticDbProvider returns a static *gorm.DB for temporary test provider overrides.
type staticDbProvider struct {
	db *gorm.DB
}

// Db returns the static database handle.
func (p staticDbProvider) Db() *gorm.DB {
	return p.db
}

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)

	// Remove temporary SQLite files before running the tests.
	fs.PurgeTestDbFiles(".", false)

	caller := "internal/entity/query/query_test.go/TestMain"
	dbc, dbn, err := testextras.AcquireDBMutex(log, caller)
	if err != nil {
		log.Error("FAIL")
		os.Exit(1)
	}
	defer testextras.UnlockDBMutex(dbc.Db())

	driver, dsn := dsn.PhotoPrismTestToDriverDSN(dbn)
	db := entity.InitTestDb(
		driver,
		dsn)

	beforeTimestamp := time.Now().UTC()
	code := m.Run()
	code = testextras.ValidateDBErrors(db.Db(), log, beforeTimestamp, code)

	testextras.ReleaseDBMutex(dbc.Db(), log, caller, code)

	// Remove temporary SQLite files after running the tests.
	db.Close()

	fs.PurgeTestDbFiles(".", false)

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

	t.Run("Postgres", func(t *testing.T) {
		if DbDialect() != Postgres {
			t.SkipNow()
		}
		assert.Equal(t, Postgres, DbDialect())
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
