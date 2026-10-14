package entity

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/photoprism/photoprism/internal/entity/sqlcount"
	"github.com/photoprism/photoprism/internal/event"
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

	var count int64

	require.NoError(t, db.Table(name).Count(&count).Error)

	return int(count)
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
		if name := UnscopedDb().Name(); name != dsn.DialectSQLite && name != dsn.DialectMySQL {
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
		if name := UnscopedDb().Name(); name != dsn.DialectMySQL {
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
		if name := UnscopedDb().Name(); name != dsn.DialectMySQL {
			t.Skipf("auto-increment tables are not looked up on %s", name)
		}

		result := autoIncrementTables(UnscopedDb())
		assert.True(t, result[Photo{}.TableName()])
		assert.True(t, result[Camera{}.TableName()])
		assert.False(t, result[Marker{}.TableName()])
		assert.False(t, result["test_truncate_missing"])
	})
	t.Run("OtherDriver", func(t *testing.T) {
		if name := UnscopedDb().Name(); name == dsn.DialectMySQL {
			t.Skipf("auto-increment tables are looked up on %s", name)
		}

		assert.Empty(t, autoIncrementTables(UnscopedDb()))
	})
}

func TestTables_Truncate(t *testing.T) {
	t.Run("KeepsSchemaTables", func(t *testing.T) {
		db := createTestTable(t, "test_truncate_list")
		versions := migrate.Version{}.TableName()
		version := migrate.NewVersion("test-truncate", "test")
		require.NoError(t, version.Create(db))
		t.Cleanup(func() { _ = db.Delete(version).Error })
		before := countTestRows(t, db, versions)
		require.NotZero(t, before, "versions must not be empty")

		Tables{"test_truncate_list": nil, versions: nil}.Truncate(db)

		assert.Equal(t, 0, countTestRows(t, db, "test_truncate_list"))
		assert.Equal(t, before, countTestRows(t, db, versions))
	})
	t.Run("RestartsAutoIncrement", func(t *testing.T) {
		if name := UnscopedDb().Name(); name != dsn.DialectMySQL {
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
	ID   uint   `gorm:"primaryKey"`
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
	t.Run("DuplicateKeys", func(t *testing.T) {
		t.Cleanup(func() { _ = Db().Migrator().DropTable(&createTableTest{}) })
		require.NoError(t, CreateTable(Db(), &createTableTest{}))
		require.NoError(t, Db().Migrator().DropIndex(&createTableTest{}, "uix_create_table_test_name"))
		require.NoError(t, Db().Exec("INSERT INTO create_table_test (name) VALUES ('a'), ('a'), ('b')").Error)

		system := captureSystemErrors(t)
		err := Tables{"create_table_test": &createTableTest{}}.Migrate(Db(), migrate.Opt(true, false, nil))
		assert.Error(t, err)
		require.Len(t, system.AllEntries(), 1)
		assert.Contains(t, system.LastEntry().Message, "unique index uix_create_table_test_name on create_table_test")
		assert.Contains(t, system.LastEntry().Message, "duplicate values (1)")
		assert.Contains(t, system.LastEntry().Message, "HAVING COUNT(*) > 1")
	})
	t.Run("ConvertDuplicateKeys", func(t *testing.T) {
		if DbDialect() != dsn.DialectSQLite {
			t.Skip("SQLite only")
		}

		t.Cleanup(func() { _ = Db().Migrator().DropTable(&createTableTest{}) })
		require.NoError(t, Db().Exec("CREATE TABLE create_table_test (id integer PRIMARY KEY AUTOINCREMENT, name varbinary(42))").Error)
		require.NoError(t, Db().Exec("INSERT INTO create_table_test (name) VALUES ('a'), ('a')").Error)

		system := captureSystemErrors(t)
		err := Tables{"create_table_test": &createTableTest{}}.Migrate(Db(), migrate.Opt(true, false, nil))
		assert.Error(t, err)
		require.Len(t, system.AllEntries(), 1)
		assert.Contains(t, system.LastEntry().Message, "unique index uix_create_table_test_name on create_table_test")
	})
}

// captureSystemErrors replaces the system logger with one that captures its error entries.
func captureSystemErrors(t *testing.T) *logtest.Hook {
	t.Helper()

	logger, hook := logtest.NewNullLogger()
	logger.SetLevel(logrus.ErrorLevel)
	previous := event.SystemLog
	event.SystemLog = logger
	t.Cleanup(func() { event.SystemLog = previous })

	return hook
}

func TestLogDuplicateKeys(t *testing.T) {
	t.Run("NoTable", func(t *testing.T) {
		system := captureSystemErrors(t)
		logDuplicateKeys(Db(), &createTableTest{})
		assert.Empty(t, system.AllEntries())
	})
	t.Run("NullValues", func(t *testing.T) {
		t.Cleanup(func() { _ = Db().Migrator().DropTable(&createTableTest{}) })
		require.NoError(t, CreateTable(Db(), &createTableTest{}))
		require.NoError(t, Db().Migrator().DropIndex(&createTableTest{}, "uix_create_table_test_name"))
		require.NoError(t, Db().Exec("INSERT INTO create_table_test (name) VALUES (NULL), (NULL), ('a')").Error)

		system := captureSystemErrors(t)
		logDuplicateKeys(Db(), &createTableTest{})
		assert.Empty(t, system.AllEntries())
	})
}

// fixtureStatement matches the ORM statements with which the default fixtures find or create their rows.
var fixtureStatement = regexp.MustCompile("(?i)^\\s*(SELECT \\* FROM|INSERT INTO)\\s+[`\"]?(auth_users|places|cells|cameras|lenses|countries)\\b")

// fixtureStatements returns the statements with which the default fixtures access their tables.
func fixtureStatements(statements []string) (result []string) {
	for _, s := range statements {
		if fixtureStatement.MatchString(s) {
			result = append(result, s)
		}
	}

	return result
}

func TestInitDb_MissingTables(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("Success", func(t *testing.T) {
		FlushCaches()
		statements := countedStatements(t, func() {
			assert.NoError(t, InitDb(migrate.Opt(false, false, nil)))
		})
		assert.NotEmpty(t, fixtureStatements(statements), "default fixtures")
	})
	t.Run("OtherTableMissing", func(t *testing.T) {
		// A registered table that does not exist after migrating fails the initialization.
		Entities["migrate_missing_test"] = &migrateFailTest{}
		t.Cleanup(func() { delete(Entities, "migrate_missing_test") })

		// The default fixtures load the stored admin account.
		admin := Admin
		require.NotEmpty(t, admin.UserUID)
		t.Cleanup(func() { Admin = admin })
		Admin.UserUID = ""

		FlushCaches()
		statements := countedStatements(t, func() {
			assert.Error(t, InitDb(migrate.Opt(false, false, nil)))
		})
		assert.NotEmpty(t, fixtureStatements(statements), "default fixtures")
		assert.Equal(t, admin.UserUID, Admin.UserUID)
	})
	t.Run("FixtureTablesMissing", func(t *testing.T) {
		dbName := filepath.Join(t.TempDir(), "missing.db")
		p, err := sqlcount.OpenGorm(dsn.DriverSQLite3, dbName, func(pool gorm.ConnPool) (*gorm.DB, error) { return OpenDbConn(dsn.DriverSQLite3, dbName, pool) })
		require.NoError(t, err)

		prev := dbConn
		SetDbProvider(p)
		t.Cleanup(func() {
			SetDbProvider(prev)
			_ = p.Close()
		})

		FlushCaches()
		p.Counter.Start()
		assert.Error(t, InitDb(migrate.Opt(false, false, nil)))
		assert.Empty(t, fixtureStatements(p.Counter.Stop()), "default fixtures")
	})
}

func TestHasDefaultFixtureTables(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert.True(t, hasDefaultFixtureTables(Db()))
	})
	t.Run("Missing", func(t *testing.T) {
		db, err := OpenDb(dsn.DriverSQLite3, filepath.Join(t.TempDir(), "fixtures.db"))
		require.NoError(t, err)
		t.Cleanup(func() { _ = CloseDb(db) })

		require.NoError(t, db.AutoMigrate(&User{}, &Place{}, &Cell{}, &Country{}, &Camera{}))
		assert.False(t, hasDefaultFixtureTables(db))

		require.NoError(t, db.AutoMigrate(&Lens{}))
		assert.True(t, hasDefaultFixtureTables(db))
	})
	t.Run("NoDb", func(t *testing.T) {
		assert.False(t, hasDefaultFixtureTables(nil))
	})
}

// createTableTest is a model for testing table creation.
type createTableTest struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"type:bytes;size:42;uniqueIndex:uix_create_table_test_name"`
}

// TableName returns the entity table name.
func (createTableTest) TableName() string {
	return "create_table_test"
}

func TestWithTableOptions(t *testing.T) {
	t.Run("Dialect", func(t *testing.T) {
		options, ok := withTableOptions(Db(), Entities, false).Get("gorm:table_options")

		if DbDialect() == dsn.DialectMySQL {
			assert.True(t, ok)
			assert.Equal(t, mysqlTableOptions(mysqlDefaultCollation), options)
		} else {
			assert.False(t, ok)
		}

		_, ok = Db().Get("gorm:table_options")
		assert.False(t, ok, "connection unchanged")

		// A chain that is not wrapped in a session must not be changed either.
		chain := Db().Model(&Photo{})
		_ = withTableOptions(chain, Entities, false)
		_, ok = chain.Get("gorm:table_options")
		assert.False(t, ok, "chain unchanged")
	})
	t.Run("SharedCollation", func(t *testing.T) {
		if DbDialect() != dsn.DialectMySQL {
			t.Skip("MySQL/MariaDB only")
		}

		t.Cleanup(func() { _ = Db().Migrator().DropTable(&createTableTest{}) })
		require.NoError(t, Db().Exec("CREATE TABLE create_table_test (id INT) ENGINE=InnoDB "+
			"DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci").Error)

		list := Tables{createTableTest{}.TableName(): &createTableTest{}}
		options, ok := withTableOptions(Db(), list, false).Get("gorm:table_options")
		assert.True(t, ok)
		assert.Equal(t, mysqlTableOptions("utf8mb4_general_ci"), options)

		// With mixed collations, the most common one is used, and the default on a tie.
		list[Photo{}.TableName()] = &Photo{}
		options, _ = withTableOptions(Db(), list, false).Get("gorm:table_options")
		assert.Equal(t, mysqlTableOptions(mysqlDefaultCollation), options)

		t.Cleanup(func() { _ = Db().Exec("DROP TABLE IF EXISTS create_table_test_2").Error })
		require.NoError(t, Db().Exec("CREATE TABLE create_table_test_2 (id INT) ENGINE=InnoDB "+
			"DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci").Error)
		list["create_table_test_2"] = &createTableTest{}
		options, _ = withTableOptions(Db(), list, false).Get("gorm:table_options")
		assert.Equal(t, mysqlTableOptions("utf8mb4_general_ci"), options)
	})
	t.Run("AutoMigrate", func(t *testing.T) {
		t.Cleanup(func() { _ = Db().Migrator().DropTable(&createTableTest{}) })
		require.NoError(t, withTableOptions(Db(), Entities, false).AutoMigrate(&createTableTest{}))
		assert.True(t, Db().Migrator().HasIndex(&createTableTest{}, "uix_create_table_test_name"))
		require.NoError(t, withTableOptions(Db(), Entities, false).AutoMigrate(&createTableTest{}))
		assert.Error(t, withTableOptions(Db(), Entities, false).AutoMigrate(&migrateFailTest{}))
	})
}

func TestCreateTable(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Cleanup(func() { _ = Db().Migrator().DropTable(&createTableTest{}) })
		statements := countStatements(t, func() { require.NoError(t, CreateTable(Db(), &createTableTest{})) })
		assert.True(t, Db().Migrator().HasTable(&createTableTest{}))
		assert.True(t, Db().Migrator().HasIndex(&createTableTest{}, "uix_create_table_test_name"))

		// MySQL tables get the table options, as the database default may differ.
		created := 0

		for _, stmt := range statements {
			if strings.HasPrefix(stmt, "CREATE TABLE") {
				created++
				assert.Equal(t, DbDialect() == dsn.DialectMySQL, strings.HasSuffix(stmt, mysqlTableOptions(mysqlDefaultCollation)), stmt)
			}
		}

		assert.Equal(t, 1, created)
	})
	t.Run("Exists", func(t *testing.T) {
		t.Cleanup(func() { _ = Db().Migrator().DropTable(&createTableTest{}) })
		require.NoError(t, CreateTable(Db(), &createTableTest{}))
		assert.Error(t, CreateTable(Db(), &createTableTest{}))
	})
}
