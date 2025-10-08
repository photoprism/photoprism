package query

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)

	// Remove temporary SQLite files before running the tests.
	fs.PurgeTestDbFiles(".", false)

	db := entity.InitTestDb(
		os.Getenv("PHOTOPRISM_TEST_DRIVER"),
		os.Getenv("PHOTOPRISM_TEST_DSN"))

	defer db.Close()

	code := m.Run()

	// Remove temporary SQLite files after running the tests.
	fs.PurgeTestDbFiles(".", false)

	os.Exit(code)
}

func TestDbDialect(t *testing.T) {
	t.Run("SQLite", func(t *testing.T) {
		assert.Equal(t, "sqlite3", DbDialect())
	})
}

func TestBatchSize(t *testing.T) {
	t.Run("SQLite", func(t *testing.T) {
		assert.Equal(t, 333, BatchSize())
	})
}
