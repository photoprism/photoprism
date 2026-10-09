package entity

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity/sqlcount"
)

func TestDbHasTable(t *testing.T) {
	t.Run("Found", func(t *testing.T) {
		found, err := DbHasTable(Db(), Photo{}.TableName())
		require.NoError(t, err)
		assert.True(t, found)
	})
	t.Run("NotFound", func(t *testing.T) {
		found, err := DbHasTable(Db(), "no_such_table")
		require.NoError(t, err)
		assert.False(t, found)
	})
	t.Run("NilDb", func(t *testing.T) {
		_, err := DbHasTable(nil, Photo{}.TableName())
		assert.Error(t, err)
	})
	t.Run("ClosedDb", func(t *testing.T) {
		db := closedTestDb(t)

		// A failing query is returned rather than raised as a panic.
		found, err := DbHasTable(db, Photo{}.TableName())
		assert.Error(t, err)
		assert.False(t, found)
	})
}

func TestDbHasColumn(t *testing.T) {
	t.Run("Found", func(t *testing.T) {
		found, err := DbHasColumn(Db(), Photo{}.TableName(), "photo_uid")
		require.NoError(t, err)
		assert.True(t, found)
	})
	t.Run("NotFound", func(t *testing.T) {
		found, err := DbHasColumn(Db(), Photo{}.TableName(), "no_such_column")
		require.NoError(t, err)
		assert.False(t, found)
	})
	t.Run("NoTable", func(t *testing.T) {
		found, err := DbHasColumn(Db(), "no_such_table", "photo_uid")
		require.NoError(t, err)
		assert.False(t, found)
	})
	t.Run("NilDb", func(t *testing.T) {
		_, err := DbHasColumn(nil, Photo{}.TableName(), "photo_uid")
		assert.Error(t, err)
	})
	t.Run("ClosedDb", func(t *testing.T) {
		db := closedTestDb(t)

		found, err := DbHasColumn(db, Photo{}.TableName(), "photo_uid")
		assert.Error(t, err)
		assert.False(t, found)
	})
}

// closedTestDb returns a connection to the test database that has already been closed.
func closedTestDb(t *testing.T) *gorm.DB {
	t.Helper()

	conn, ok := dbConn.(*DbConn)
	require.True(t, ok, "test database provider")

	p, err := sqlcount.OpenGorm(conn.Driver, conn.Dsn)
	require.NoError(t, err)
	require.NoError(t, p.Close())

	return p.DB
}
