package entity

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/photoprism/photoprism/pkg/dsn"
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

// TestTruncateTable checks that rows are removed, within the caller's transaction where no counter is restarted.
func TestTruncateTable(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db := createTestTable(t, "test_truncate")
		assert.Equal(t, 2, countTestRows(t, db, "test_truncate"))
		assert.NoError(t, truncateTable(db, "test_truncate", false))
		assert.Equal(t, 0, countTestRows(t, db, "test_truncate"))
	})
	t.Run("UnknownTable", func(t *testing.T) {
		assert.Error(t, truncateTable(UnscopedDb(), "test_truncate_missing", false))
	})
	t.Run("Transactional", func(t *testing.T) {
		// A DELETE is rolled back with its transaction, while TRUNCATE commits implicitly on MySQL/MariaDB.
		if name := UnscopedDb().Dialect().GetName(); name != dsn.DriverSQLite3 && name != dsn.DriverMySQL {
			t.Skipf("rows are truncated on %s", name)
		}

		db := createTestTable(t, "test_truncate_tx")
		tx := db.Begin()
		require.NoError(t, tx.Error)
		t.Cleanup(func() { tx.Rollback() })
		require.NoError(t, truncateTable(tx, "test_truncate_tx", false))
		assert.Equal(t, 0, countTestRows(t, tx, "test_truncate_tx"))
		require.NoError(t, tx.Rollback().Error)
		assert.Equal(t, 2, countTestRows(t, db, "test_truncate_tx"))
	})
	t.Run("AutoIncrement", func(t *testing.T) {
		if name := UnscopedDb().Dialect().GetName(); name != dsn.DriverMySQL {
			t.Skipf("auto-increment counters are not restarted on %s", name)
		}

		db := UnscopedDb()
		require.NoError(t, db.Exec("CREATE TABLE test_truncate_ai (id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY, test_name VARCHAR(16))").Error)
		t.Cleanup(func() { assert.NoError(t, db.Exec("DROP TABLE test_truncate_ai").Error) })
		require.NoError(t, db.Exec("INSERT INTO test_truncate_ai (id, test_name) VALUES (1000, 'foo')").Error)

		require.NoError(t, truncateTable(db, "test_truncate_ai", true))
		require.NoError(t, db.Exec("INSERT INTO test_truncate_ai (test_name) VALUES ('bar')").Error)

		var id int
		require.NoError(t, db.Raw("SELECT id FROM test_truncate_ai").Row().Scan(&id))
		assert.Equal(t, 1, id)
	})
}

// TestAutoIncrementTables checks that only tables with an auto-increment column are returned on MySQL/MariaDB.
func TestAutoIncrementTables(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		if name := UnscopedDb().Dialect().GetName(); name != dsn.DriverMySQL {
			t.Skipf("auto-increment tables are not looked up on %s", name)
		}

		result := autoIncrementTables(UnscopedDb())
		assert.True(t, result[Photo{}.TableName()])
		assert.True(t, result[Camera{}.TableName()])
		assert.False(t, result[Marker{}.TableName()])
		assert.False(t, result["test_truncate_missing"])
	})
	t.Run("OtherDriver", func(t *testing.T) {
		if name := UnscopedDb().Dialect().GetName(); name == dsn.DriverMySQL {
			t.Skipf("auto-increment tables are looked up on %s", name)
		}

		assert.Empty(t, autoIncrementTables(UnscopedDb()))
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
	t.Run("RestartsAutoIncrement", func(t *testing.T) {
		if name := UnscopedDb().Dialect().GetName(); name != dsn.DriverMySQL {
			t.Skipf("auto-increment counters are not restarted on %s", name)
		}

		db := UnscopedDb()
		require.NoError(t, db.Exec("CREATE TABLE test_truncate_list_ai (id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY, test_name VARCHAR(16))").Error)
		t.Cleanup(func() { assert.NoError(t, db.Exec("DROP TABLE test_truncate_list_ai").Error) })
		require.NoError(t, db.Exec("INSERT INTO test_truncate_list_ai (id, test_name) VALUES (1000, 'foo')").Error)

		Tables{"test_truncate_list_ai": nil}.Truncate(db)
		require.NoError(t, db.Exec("INSERT INTO test_truncate_list_ai (test_name) VALUES ('bar')").Error)

		var id int
		require.NoError(t, db.Raw("SELECT id FROM test_truncate_list_ai").Row().Scan(&id))
		assert.Equal(t, 1, id)
	})
}

// migrateFailTest is a model whose table cannot be created, to test how migration failures are reported.
type migrateFailTest struct {
	ID   uint   `gorm:"primary_key"`
	Name string `gorm:"type:BROKEN((("`
}

// TableName returns the entity table name.
func (migrateFailTest) TableName() string {
	return "migrate_fail_test"
}

func TestTables_Migrate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert.NoError(t, Tables{"photos": &Photo{}}.Migrate(Db(), migrate.Opt(true, false, nil)))
	})
	t.Run("AutoMigrateFails", func(t *testing.T) {
		err := Tables{"migrate_fail_test": &migrateFailTest{}}.Migrate(Db(), migrate.Opt(true, false, nil))
		assert.Error(t, err)
	})
}

func TestInitDb_MissingTables(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// A registered table that does not exist after migrating fails the initialization.
	Entities["migrate_missing_test"] = &migrateFailTest{}
	t.Cleanup(func() { delete(Entities, "migrate_missing_test") })

	assert.Error(t, InitDb(migrate.Opt(false, false, nil)))
}
