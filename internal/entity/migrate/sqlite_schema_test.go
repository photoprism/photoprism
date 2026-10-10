package migrate

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// convertTest is a model whose legacy table is converted in tests.
type convertTest struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"type:bytes;size:42;uniqueIndex:uix_convert_tests_name"`
	Title string `gorm:"size:100"`
	Flag  bool
}

// TableName returns the entity table name.
func (convertTest) TableName() string {
	return "convert_tests"
}

// legacyConvertDb returns a SQLite database with a legacy convert_tests table and rows.
func legacyConvertDb(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := openTestDb("sqlite3", filepath.Join(t.TempDir(), "convert.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = closeTestDb(db) })

	for _, stmt := range []string{
		"CREATE TABLE convert_tests (id integer primary key autoincrement, name VARBINARY(42), Title VARCHAR(100), flag bool, extra VARCHAR(10))",
		"CREATE UNIQUE INDEX uix_convert_tests_name ON convert_tests(name)",
		"CREATE INDEX idx_convert_tests_title ON convert_tests(title)",
		"INSERT INTO convert_tests (id, name, title, flag, extra) VALUES (1, '2020', 'First', 1, '0123')",
		"INSERT INTO convert_tests (id, name, title, flag, extra) VALUES (2, 'second', 'Second', 0, NULL)",
		"INSERT INTO convert_tests (id, name, title, flag, extra) VALUES (7, 'deleted', 'Deleted', 0, NULL)",
		"DELETE FROM convert_tests WHERE id = 7",
	} {
		require.NoError(t, db.Exec(stmt).Error, stmt)
	}

	return db
}

// tableSQL returns the CREATE TABLE statement of a SQLite table.
func tableSQL(t *testing.T, db *gorm.DB, name string) (ddl string) {
	t.Helper()
	require.NoError(t, db.Raw("SELECT sql FROM sqlite_master WHERE type = 'table' AND name = ?", name).Row().Scan(&ddl))
	return ddl
}

func TestSqliteColumnType(t *testing.T) {
	assert.Equal(t, "blob(0)", sqliteColumnType("VARBINARY(42)"))
	assert.Equal(t, "blob", sqliteColumnType("mediumblob"))
	assert.Equal(t, "varchar(255)", sqliteColumnType("varchar(255)"))
	assert.Equal(t, "text", sqliteColumnType("LONGTEXT"))
	assert.Equal(t, "numeric", sqliteColumnType("bool"))
	assert.Equal(t, "integer", sqliteColumnType("integer"))
	assert.Equal(t, "DATETIME", sqliteColumnType("DATETIME"))
}

func TestConvertSQLiteSchema(t *testing.T) {
	models := map[string]any{"convert_tests": &convertTest{}}

	t.Run("NoDatabase", func(t *testing.T) {
		assert.Error(t, ConvertSQLiteSchema(nil, models))
	})
	t.Run("MissingTable", func(t *testing.T) {
		db, err := openTestDb("sqlite3", filepath.Join(t.TempDir(), "missing.db"))
		require.NoError(t, err)
		t.Cleanup(func() { _ = closeTestDb(db) })
		require.NoError(t, ConvertSQLiteSchema(db, models))
		assert.False(t, db.Migrator().HasTable("convert_tests"))
	})
	t.Run("Success", func(t *testing.T) {
		db := legacyConvertDb(t)
		require.NoError(t, ConvertSQLiteSchema(db, models))

		ddl := tableSQL(t, db, "convert_tests")
		assert.False(t, sqliteLegacyTypes.MatchString(ddl), ddl)
		assert.False(t, db.Migrator().HasTable("convert_tests__v1"))
		assert.True(t, db.Migrator().HasIndex(&convertTest{}, "uix_convert_tests_name"))
		assert.True(t, db.Migrator().HasIndex(&convertTest{}, "idx_convert_tests_title"), "index added by a migration")

		// Numbers stored in text columns become text, columns without a field are kept.
		var rows []struct {
			ID        uint
			Name      string
			NameType  string
			Extra     *string
			ExtraType string
			Flag      bool
		}
		require.NoError(t, db.Raw("SELECT id, name, typeof(name) AS name_type, extra, typeof(extra) AS extra_type, flag FROM convert_tests ORDER BY id").Scan(&rows).Error)
		require.Len(t, rows, 2)

		var title string
		require.NoError(t, db.Raw("SELECT title FROM convert_tests WHERE id = 1").Row().Scan(&title))
		assert.Equal(t, "First", title, "column name compared case-insensitively")
		assert.Equal(t, "2020", rows[0].Name)
		assert.Equal(t, "text", rows[0].NameType)
		require.NotNil(t, rows[0].Extra)
		assert.Equal(t, "0123", *rows[0].Extra)
		assert.Equal(t, "text", rows[0].ExtraType)
		assert.True(t, rows[0].Flag)
		assert.Equal(t, "second", rows[1].Name)
		assert.Nil(t, rows[1].Extra)
		assert.False(t, rows[1].Flag)

		// The counter of the key is kept, so the ID of the deleted row is not used again.
		var seq int
		require.NoError(t, db.Raw("SELECT seq FROM sqlite_sequence WHERE name = 'convert_tests'").Row().Scan(&seq))
		assert.Equal(t, 7, seq)
		row := convertTest{Name: "third"}
		require.NoError(t, db.Create(&row).Error)
		assert.Equal(t, uint(8), row.ID)
	})
	t.Run("Converted", func(t *testing.T) {
		db := legacyConvertDb(t)
		require.NoError(t, ConvertSQLiteSchema(db, models))
		ddl := tableSQL(t, db, "convert_tests")
		require.NoError(t, ConvertSQLiteSchema(db, models))
		assert.Equal(t, ddl, tableSQL(t, db, "convert_tests"))
	})
}

func TestConvertSQLiteTable(t *testing.T) {
	t.Run("RollsBack", func(t *testing.T) {
		db := legacyConvertDb(t)

		// A duplicate in the unique column fails the copy, so the legacy table must stay as it was.
		require.NoError(t, db.Exec("DROP INDEX uix_convert_tests_name").Error)
		require.NoError(t, db.Exec("INSERT INTO convert_tests (id, name, title) VALUES (3, 'second', 'Duplicate')").Error)
		ddl := tableSQL(t, db, "convert_tests")

		assert.Error(t, convertSQLiteTable(db, "convert_tests", &convertTest{}))
		assert.Equal(t, ddl, tableSQL(t, db, "convert_tests"))
		assert.False(t, db.Migrator().HasTable("convert_tests__v1"))

		var count int64
		require.NoError(t, db.Table("convert_tests").Count(&count).Error)
		assert.Equal(t, int64(3), count)
	})
}

func TestSqliteIdent(t *testing.T) {
	assert.Equal(t, `"photos"`, sqliteIdent("photos"))
	assert.Equal(t, `"a""b"`, sqliteIdent(`a"b`))
	assert.Equal(t, `"a\b"`, sqliteIdent(`a\b`))
	assert.Equal(t, `""`, sqliteIdent(""))
}
