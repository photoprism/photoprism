package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/dsn"
)

// TestConfig_OpenTestDb checks that setup can precede global connection publication.
func TestConfig_OpenTestDb(t *testing.T) {
	previous := entity.Db()
	conf := NewIsolatedTestConfig("opentestdb", t.TempDir(), true)
	if conf.DatabaseDriver() == dsn.DriverSQLite3 {
		conf.options.DatabaseDSN = filepath.Join(conf.StoragePath(), "index.db")
	}
	db, err := conf.OpenTestDb()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	require.NoError(t, db.DB().Ping())
	assert.Same(t, previous, entity.Db())
	assert.Same(t, db, conf.Db())
	again, err := conf.OpenTestDb()
	require.NoError(t, err)
	assert.Same(t, db, again)
	assert.Same(t, previous, entity.Db())
}
