package migrate

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// tidbDialector is a SQLite dialector that reports the name of a driver without migrations.
type tidbDialector struct {
	*sqlite.Dialector
}

// Name returns the name of the driver.
func (tidbDialector) Name() string {
	return "tidb"
}

// openUnsupportedDb opens a SQLite database whose dialector reports an unsupported driver.
func openUnsupportedDb(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(tidbDialector{Dialector: &sqlite.Dialector{DSN: filepath.Join(t.TempDir(), "tidb.db")}}, &gorm.Config{})
	require.NoError(t, err)
	t.Cleanup(func() {
		if sqlDb, dbErr := db.DB(); dbErr == nil {
			_ = sqlDb.Close()
		}
	})

	return db
}

func TestRun(t *testing.T) {
	t.Run("NoDb", func(t *testing.T) {
		assert.Error(t, Run(nil, Opt(true, false, nil)))
	})
	t.Run("UnsupportedDialect", func(t *testing.T) {
		err := Run(openUnsupportedDb(t), Opt(true, false, nil))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported sql dialect")
	})
}

func TestStatus(t *testing.T) {
	t.Run("NoDb", func(t *testing.T) {
		_, err := Status(nil, nil)
		assert.Error(t, err)
	})
	t.Run("UnsupportedDialect", func(t *testing.T) {
		_, err := Status(openUnsupportedDb(t), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported sql dialect")
	})
}

func TestDialects(t *testing.T) {
	for name := range Dialects {
		assert.NotNil(t, once[name], name)
	}

	assert.Empty(t, DialectPostgreSQL)
}
