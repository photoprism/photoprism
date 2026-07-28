package entity

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity/migrate"
)

// createTestTable creates a table with two rows and removes it on test cleanup.
func createTestTable(t *testing.T, name string) *gorm.DB {
	t.Helper()

	db := UnscopedDb()

	require.NoError(t, db.Exec("CREATE TABLE "+name+" (id INTEGER PRIMARY KEY, test_name VARCHAR(16))").Error)

	t.Cleanup(func() {
		assert.NoError(t, db.Exec("DROP TABLE "+name).Error)
	})

	require.NoError(t, db.Exec("INSERT INTO "+name+" (id, test_name) VALUES (1, 'foo'), (2, 'bar')").Error)

	return db
}

// countTestRows returns the number of rows in the table with the specified name.
func countTestRows(t *testing.T, db *gorm.DB, name string) int {
	t.Helper()

	var count int

	require.NoError(t, db.Table(name).Count(&count).Error)

	return count
}

func TestTruncateTable(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db := createTestTable(t, "test_truncate")
		assert.Equal(t, 2, countTestRows(t, db, "test_truncate"))
		assert.NoError(t, truncateTable(db, "test_truncate"))
		assert.Equal(t, 0, countTestRows(t, db, "test_truncate"))
	})
	t.Run("UnknownTable", func(t *testing.T) {
		assert.Error(t, truncateTable(UnscopedDb(), "test_truncate_missing"))
	})
}

func TestTables_Truncate(t *testing.T) {
	t.Run("KeepsSchemaTables", func(t *testing.T) {
		db := createTestTable(t, "test_truncate_list")
		versions := migrate.Version{}.TableName()
		before := countTestRows(t, db, versions)
		require.NotZero(t, before, "versions must not be empty")

		Tables{"test_truncate_list": nil, versions: nil}.Truncate(db)

		assert.Equal(t, 0, countTestRows(t, db, "test_truncate_list"))
		assert.Equal(t, before, countTestRows(t, db, versions))
	})
}
