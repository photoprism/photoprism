package entity

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Servers of the development environment that transfer tests use as targets, unless the environment names others.
const (
	transferTestMariaDbDSN  = "root:photoprism@tcp(mariadb:4001)/testdb?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true"
	transferTestPostgresDSN = "postgres://photoprism:photoprism@postgres.:4002/testdb?sslmode=disable&TimeZone=UTC" //nolint:gosec // development environment
)

// transferTestServerDSN returns the DSN of the server that a transfer test may create a target database on: the test
// database's own if it uses the driver, otherwise PHOTOPRISM_TEST_DSN_<DRIVER> or the development environment's service.
func transferTestServerDSN(driver string) string {
	if dsn.ParseDriver(os.Getenv("PHOTOPRISM_TEST_DRIVER")) == driver {
		return os.Getenv("PHOTOPRISM_TEST_DSN")
	}

	switch driver {
	case dsn.DriverMySQL:
		if s := os.Getenv("PHOTOPRISM_TEST_DSN_MARIADB"); s != "" {
			return s
		}

		return transferTestMariaDbDSN
	case dsn.DriverPostgres:
		if s := os.Getenv("PHOTOPRISM_TEST_DSN_POSTGRES"); s != "" {
			return s
		}

		return transferTestPostgresDSN
	}

	return ""
}

// transferTestTarget opens an empty database of the specified driver with a migrated schema, and skips the
// test if its server cannot be reached. The database is removed when the test ends.
func transferTestTarget(t *testing.T, driver string) *gorm.DB {
	t.Helper()

	var targetDsn string

	switch driver {
	case dsn.DriverSQLite3:
		targetDsn = fmt.Sprintf("%s?%s", filepath.Join(t.TempDir(), "target.db"), dsn.Params[dsn.DriverSQLite3])
	case dsn.DriverMySQL:
		conf, err := mysql.ParseDSN(transferTestServerDSN(driver))
		require.NoError(t, err)
		skipUnlessReachable(t, conf.Addr, "mariadb")
		name := testDbName("testdb_transfer", testDbNameLen)
		require.NoError(t, createTestDb(conf, name))
		t.Cleanup(func() { dropTransferTestDb(t, driver, conf.FormatDSN(), name) })
		conf.DBName = name
		targetDsn = conf.FormatDSN()
	case dsn.DriverPostgres:
		serverDsn := transferTestServerDSN(driver)
		u, err := url.Parse(serverDsn)
		require.NoError(t, err)
		skipUnlessReachable(t, u.Host, "postgres")
		name := testDbName("testdb_transfer", testPostgresDbNameLen)
		dropTransferTestDb(t, driver, serverDsn, name)
		require.NoError(t, createPostgresTestDb(serverDsn, name))
		t.Cleanup(func() { dropTransferTestDb(t, driver, serverDsn, name) })
		u.Path = "/" + name
		targetDsn = u.String()
	default:
		t.Fatalf("unsupported driver %s", driver)
	}

	db, err := OpenDb(driver, targetDsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = CloseDb(db) })

	if driver == dsn.DriverMySQL {
		// A database that was not removed after an earlier run starts empty, as in production.
		for name := range Entities {
			require.NoError(t, db.Migrator().DropTable(name))
		}
	}

	require.NoError(t, db.AutoMigrate(&migrate.Migration{}, &migrate.Version{}))
	require.NoError(t, Entities.Migrate(db, migrate.Opt(true, false, nil)))

	return db
}

// skipUnlessReachable skips the test if nothing listens on addr.
func skipUnlessReachable(t *testing.T, addr, service string) {
	t.Helper()

	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)

	if err != nil {
		t.Skipf("%s service not reachable at %s, start it with docker compose to run this test", service, addr)
	}

	_ = conn.Close()
}

// dropTransferTestDb removes a database created by transferTestTarget, whose name testDbName restricts to identifier characters.
func dropTransferTestDb(t *testing.T, driver, serverDsn, name string) {
	t.Helper()

	var db *sql.DB

	switch driver {
	case dsn.DriverMySQL:
		conf, err := mysql.ParseDSN(serverDsn)
		require.NoError(t, err)
		conf.DBName = ""
		db, err = sql.Open(dsn.DriverMySQL, conf.FormatDSN())
		require.NoError(t, err)
	case dsn.DriverPostgres:
		conf, err := pgx.ParseConfig(serverDsn)
		require.NoError(t, err)
		db = stdlib.OpenDB(*conf)
	}

	defer db.Close()

	_, err := db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", name))
	assert.NoError(t, err)
}

// transferTestRows returns all rows of a table as field values ordered by primary key, with times in UTC and
// whole seconds, the precision that every database stores.
func transferTestRows(t *testing.T, db *gorm.DB, name string, model any) [][]any {
	t.Helper()

	stmt := &gorm.Statement{DB: db}
	require.NoError(t, stmt.Parse(model))

	order := clause.OrderBy{}

	for _, f := range stmt.Schema.PrimaryFields {
		order.Columns = append(order.Columns, clause.OrderByColumn{Column: clause.Column{Name: f.DBName}})
	}

	rows := reflect.New(reflect.SliceOf(reflect.PointerTo(stmt.Schema.ModelType)))
	require.NoError(t, db.Session(&gorm.Session{NewDB: true, SkipHooks: true}).Table(name).Unscoped().Order(order).Find(rows.Interface()).Error)

	result := make([][]any, rows.Elem().Len())

	for i := range result {
		rv := rows.Elem().Index(i).Elem()

		for _, dbName := range stmt.Schema.DBNames {
			v := rv.FieldByIndex(stmt.Schema.FieldsByDBName[dbName].StructField.Index).Interface()

			switch tv := v.(type) {
			case time.Time:
				v = tv.UTC().Truncate(time.Second)
			case *time.Time:
				if tv != nil {
					v = tv.UTC().Truncate(time.Second)
				}
			}

			result[i] = append(result[i], v)
		}
	}

	return result
}

func TestTables_DataTables(t *testing.T) {
	names := Entities.DataTables()

	assert.NotContains(t, names, migrate.Migration{}.TableName())
	assert.NotContains(t, names, migrate.Version{}.TableName())
	assert.Contains(t, names, Photo{}.TableName())
	assert.Len(t, names, len(Entities)-2)
	assert.IsIncreasing(t, names)
}

func TestTables_Rows(t *testing.T) {
	t.Run("TestDatabase", func(t *testing.T) {
		rows, err := Entities.Rows(Db())

		require.NoError(t, err)
		assert.Positive(t, rows[Photo{}.TableName()])
		assert.NotContains(t, rows, migrate.Version{}.TableName())
	})
	t.Run("NoTables", func(t *testing.T) {
		db, err := OpenDb(dsn.DriverSQLite3, filepath.Join(t.TempDir(), "empty.db"))
		require.NoError(t, err)
		t.Cleanup(func() { _ = CloseDb(db) })

		rows, err := Entities.Rows(db)

		require.NoError(t, err)
		assert.Empty(t, rows)
	})
	t.Run("EmptyTables", func(t *testing.T) {
		rows, err := Entities.Rows(transferTestTarget(t, dsn.DriverSQLite3))

		require.NoError(t, err)
		assert.Empty(t, rows)
	})
}

func TestTransferKeyset(t *testing.T) {
	a := &schema.Field{DBName: "a"}
	b := &schema.Field{DBName: "b"}

	t.Run("OneKey", func(t *testing.T) {
		cond, args := transferKeyset([]*schema.Field{a}, []any{1})

		assert.Equal(t, "(? > ?)", cond)
		assert.Equal(t, []any{clause.Column{Name: "a"}, 1}, args)
	})
	t.Run("MissingValues", func(t *testing.T) {
		cond, args := transferKeyset([]*schema.Field{a, b}, []any{"x"})

		assert.Equal(t, "1 = 0", cond)
		assert.Empty(t, args)
	})
	t.Run("TwoKeys", func(t *testing.T) {
		cond, args := transferKeyset([]*schema.Field{a, b}, []any{"x", "y"})

		assert.Equal(t, "(? >= ?) AND ((? > ?) OR (? = ? AND ? > ?))", cond)
		assert.Equal(t, []any{
			clause.Column{Name: "a"}, "x",
			clause.Column{Name: "a"}, "x",
			clause.Column{Name: "a"}, "x", clause.Column{Name: "b"}, "y",
		}, args)
	})
}

func TestSameDatabase(t *testing.T) {
	t.Run("SameConnection", func(t *testing.T) {
		same, err := SameDatabase(Db(), Db())

		require.NoError(t, err)
		assert.True(t, same)
	})
	t.Run("SameFile", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "same.db")
		a, err := OpenDb(dsn.DriverSQLite3, fileName)
		require.NoError(t, err)
		t.Cleanup(func() { _ = CloseDb(a) })
		b, err := OpenDb(dsn.DriverSQLite3, fileName)
		require.NoError(t, err)
		t.Cleanup(func() { _ = CloseDb(b) })

		same, err := SameDatabase(a, b)

		require.NoError(t, err)
		assert.True(t, same)
		assert.Empty(t, transferProbeTables(t, b))
	})
	t.Run("OtherDatabase", func(t *testing.T) {
		driver := dsn.ParseDriver(Db().Name())
		target := transferTestTarget(t, driver)

		same, err := SameDatabase(Db(), target)

		require.NoError(t, err)
		assert.False(t, same)
		assert.Empty(t, transferProbeTables(t, target))
	})
	t.Run("OtherDialect", func(t *testing.T) {
		driver := dsn.DriverSQLite3

		if dsn.ParseDriver(Db().Name()) == dsn.DriverSQLite3 {
			driver = dsn.DriverPostgres
		}

		same, err := SameDatabase(Db(), transferTestTarget(t, driver))

		require.NoError(t, err)
		assert.False(t, same)
	})
	t.Run("NotConnected", func(t *testing.T) {
		_, err := SameDatabase(nil, Db())

		assert.Error(t, err)
	})
}

// transferProbeTables returns the names of the probe tables that remain in db.
func transferProbeTables(t *testing.T, db *gorm.DB) (result []string) {
	t.Helper()

	tables, err := db.Migrator().GetTables()
	require.NoError(t, err)

	for _, name := range tables {
		if strings.HasPrefix(name, "transfer_probe_") {
			result = append(result, name)
		}
	}

	return result
}

// TestTables_TransferSubset copies two tables into a new SQLite file, so short runs keep an end-to-end
// transfer while TestTables_Transfer covers every table and driver.
func TestTables_TransferSubset(t *testing.T) {
	albums, keywords := Album{}.TableName(), Keyword{}.TableName()
	subset := Tables{albums: Entities[albums], keywords: Entities[keywords]}

	dst, err := OpenDb(dsn.DriverSQLite3, fmt.Sprintf("%s?%s", filepath.Join(t.TempDir(), "target.db"), dsn.Params[dsn.DriverSQLite3]))
	require.NoError(t, err)
	t.Cleanup(func() { _ = CloseDb(dst) })
	require.NoError(t, dst.AutoMigrate(Entities[albums], Entities[keywords]))

	result, err := subset.Transfer(Db(), dst, TransferOptions{BatchSize: 7})
	require.NoError(t, err)

	for _, name := range []string{albums, keywords} {
		src := transferTestRows(t, Db(), name, Entities[name])

		assert.NotEmpty(t, src, name)
		assert.Len(t, src, int(result[name]), name)
		assert.Equal(t, src, transferTestRows(t, dst, name, Entities[name]), name)
	}
}

func TestTables_Transfer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// Rows whose values a default or a numeric conversion would change.
	album := &Album{AlbumUID: rnd.GenerateUID(AlbumUID), AlbumSlug: "transfer-test", AlbumTitle: "Transfer Test", AlbumType: AlbumManual}
	require.NoError(t, UnscopedDb().Create(album).Error)
	require.NoError(t, UnscopedDb().Model(album).UpdateColumns(Values{"album_country": "", "album_type": ""}).Error)
	keywords := []*Keyword{{Keyword: "0123"}, {Keyword: "123"}}
	require.NoError(t, UnscopedDb().Create(&keywords).Error)
	marker := MarkerFixtures.Get("1000003-4")
	require.NoError(t, UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", marker.MarkerUID).UpdateColumn("landmarks_json", []byte{}).Error)

	t.Cleanup(func() {
		UnscopedDb().Delete(album)
		UnscopedDb().Delete(keywords[0])
		UnscopedDb().Delete(keywords[1])
		UnscopedDb().Model(&Marker{}).Where("marker_uid = ?", marker.MarkerUID).UpdateColumn("landmarks_json", []byte(marker.LandmarksJSON))
	})

	for _, driver := range []string{dsn.DriverSQLite3, dsn.DriverMySQL, dsn.DriverPostgres} {
		t.Run(driver, func(t *testing.T) {
			dst := transferTestTarget(t, driver)

			// Rows in the target are replaced.
			require.NoError(t, dst.Create(&Error{ErrorMessage: "transfer test"}).Error)

			var progress []string

			result, err := Entities.Transfer(Db(), dst, TransferOptions{
				BatchSize: 7,
				Progress:  func(table string, rows int64) { progress = append(progress, table) },
			})

			require.NoError(t, err)
			assert.Equal(t, Entities.DataTables(), progress)

			for _, name := range Entities.DataTables() {
				if name == (Error{}).TableName() {
					// The source logs errors while the test runs.
					continue
				}

				src := transferTestRows(t, Db(), name, Entities[name])

				assert.Len(t, src, int(result[name]), name)
				assert.Equal(t, src, transferTestRows(t, dst, name, Entities[name]), name)
			}

			var found Album
			require.NoError(t, dst.Table(Album{}.TableName()).Where("album_uid = ?", album.AlbumUID).Take(&found).Error)
			assert.Equal(t, "", found.AlbumCountry)
			assert.Equal(t, "", found.AlbumType)

			var names []string
			require.NoError(t, dst.Model(&Keyword{}).Where("keyword IN (?)", []string{"0123", "123"}).Order("keyword").Pluck("keyword", &names).Error)
			assert.Equal(t, []string{"0123", "123"}, names)

			var emptyJson int64
			require.NoError(t, dst.Model(&Marker{}).Where("marker_uid = ? AND landmarks_json IS NOT NULL AND LENGTH(landmarks_json) = 0", marker.MarkerUID).Count(&emptyJson).Error)
			assert.Equal(t, int64(1), emptyJson)

			var messages []string
			require.NoError(t, dst.Model(&Error{}).Where("error_message = ?", "transfer test").Pluck("error_message", &messages).Error)
			assert.Empty(t, messages)

			// New rows get IDs after the copied ones.
			lens := &Lens{LensSlug: "transfer-test-" + driver, LensName: "Transfer Test"}
			require.NoError(t, dst.Create(lens).Error)

			var maxID uint
			require.NoError(t, Db().Model(&Lens{}).Select("MAX(id)").Row().Scan(&maxID))
			assert.Greater(t, lens.ID, maxID)
		})
	}
	t.Run("NotConnected", func(t *testing.T) {
		_, err := Entities.Transfer(Db(), nil, TransferOptions{})

		assert.Error(t, err)
	})
	t.Run("FailedInsert", func(t *testing.T) {
		list := Tables{Camera{}.TableName(): &Camera{}, Lens{}.TableName(): &Lens{}}
		dst := transferTestTarget(t, dsn.DriverSQLite3)

		// The target cannot store lenses, which are copied after cameras.
		require.NoError(t, dst.Migrator().DropTable(Lens{}.TableName()))
		require.NoError(t, dst.Exec("CREATE TABLE lenses (id INTEGER PRIMARY KEY)").Error)

		_, err := list.Transfer(Db(), dst, TransferOptions{})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "copy lenses")
		assert.Contains(t, err.Error(), "rows 1 to")

		var cameras int64
		require.NoError(t, dst.Model(&Camera{}).Count(&cameras).Error)
		assert.Zero(t, cameras)
	})
	t.Run("SkippedRows", func(t *testing.T) {
		src, err := OpenDb(dsn.DriverSQLite3, filepath.Join(t.TempDir(), "source.db"))
		require.NoError(t, err)
		t.Cleanup(func() { _ = CloseDb(src) })
		dst, err := OpenDb(dsn.DriverSQLite3, filepath.Join(t.TempDir(), "target.db"))
		require.NoError(t, err)
		t.Cleanup(func() { _ = CloseDb(dst) })

		// SQLite allows NULL in a composite primary key, so keyset paging cannot reach the second row.
		for _, db := range []*gorm.DB{src, dst} {
			require.NoError(t, db.Exec("CREATE TABLE transfer_nulls (a TEXT, b TEXT, PRIMARY KEY (a, b))").Error)
		}

		require.NoError(t, src.Exec("INSERT INTO transfer_nulls (a, b) VALUES ('x', NULL), ('x', NULL)").Error)

		_, err = Tables{"transfer_nulls": &transferTestNull{}}.Transfer(src, dst, TransferOptions{BatchSize: 1})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "1 of 2 rows copied from transfer_nulls")
	})
	t.Run("ChangedSource", func(t *testing.T) {
		// In WAL mode, SQLite lets a writer commit while the transfer holds its read snapshot.
		src, err := OpenDb(dsn.DriverSQLite3, filepath.Join(t.TempDir(), "source.db")+"?_journal_mode=WAL")
		require.NoError(t, err)
		t.Cleanup(func() { _ = CloseDb(src) })
		require.NoError(t, src.AutoMigrate(&Lens{}))
		require.NoError(t, src.Create(&Lens{LensSlug: "a", LensName: "A"}).Error)
		dst := transferTestTarget(t, dsn.DriverSQLite3)

		// A row added after the copy, e.g. by an instance that still runs, fails the transfer.
		_, err = Tables{Lens{}.TableName(): &Lens{}}.Transfer(src, dst, TransferOptions{
			Progress: func(string, int64) { require.NoError(t, src.Create(&Lens{LensSlug: "b", LensName: "B"}).Error) },
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "the index changed during the transfer (lenses)")

		var lenses int64
		require.NoError(t, dst.Model(&Lens{}).Count(&lenses).Error)
		assert.Zero(t, lenses)
	})
	t.Run("UpdatedSource", func(t *testing.T) {
		src, err := OpenDb(dsn.DriverSQLite3, filepath.Join(t.TempDir(), "source.db")+"?_journal_mode=WAL")
		require.NoError(t, err)
		t.Cleanup(func() { _ = CloseDb(src) })
		require.NoError(t, src.AutoMigrate(&Lens{}))
		lens := &Lens{LensSlug: "a", LensName: "A"}
		require.NoError(t, src.Create(lens).Error)
		dst := transferTestTarget(t, dsn.DriverSQLite3)

		// An update after the copy leaves the number of rows unchanged, but not the latest update time.
		_, err = Tables{Lens{}.TableName(): &Lens{}}.Transfer(src, dst, TransferOptions{
			Progress: func(string, int64) {
				require.NoError(t, src.Model(lens).Updates(Values{"lens_name": "B", "updated_at": lens.UpdatedAt.Add(time.Hour)}).Error)
			},
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "the index changed during the transfer (lenses)")
	})
	t.Run("LogTable", func(t *testing.T) {
		src, err := OpenDb(dsn.DriverSQLite3, filepath.Join(t.TempDir(), "source.db")+"?_journal_mode=WAL")
		require.NoError(t, err)
		t.Cleanup(func() { _ = CloseDb(src) })
		require.NoError(t, src.AutoMigrate(&Error{}))
		dst := transferTestTarget(t, dsn.DriverSQLite3)

		// Entries in a log table, e.g. from the transfer itself, do not fail it.
		result, err := Tables{Error{}.TableName(): &Error{}}.Transfer(src, dst, TransferOptions{
			Progress: func(string, int64) { require.NoError(t, src.Create(&Error{ErrorMessage: "late"}).Error) },
		})

		require.NoError(t, err)
		assert.Zero(t, result[Error{}.TableName()])
	})
	t.Run("SkipDuplicates", func(t *testing.T) {
		for _, driver := range []string{dsn.DriverSQLite3, dsn.DriverMySQL, dsn.DriverPostgres} {
			t.Run(driver, func(t *testing.T) {
				src, err := OpenDb(dsn.DriverSQLite3, filepath.Join(t.TempDir(), "source.db"))
				require.NoError(t, err)
				t.Cleanup(func() { _ = CloseDb(src) })
				dst := transferTestTarget(t, driver)

				// The target compares names case-insensitively, the source byte by byte.
				require.NoError(t, src.Exec("CREATE TABLE transfer_logins (id INTEGER PRIMARY KEY, name VARCHAR(64))").Error)
				require.NoError(t, src.Exec("INSERT INTO transfer_logins (id, name) VALUES (1, 'admin'), (2, 'other'), (3, 'Other'), (4, 'more'), (5, 'ADMIN'), (6, 'last')").Error)

				switch driver {
				case dsn.DriverMySQL:
					require.NoError(t, dst.Exec("CREATE TABLE transfer_logins (id INTEGER PRIMARY KEY, name VARCHAR(64) NOT NULL COLLATE utf8mb4_unicode_ci UNIQUE)").Error)
				case dsn.DriverPostgres:
					require.NoError(t, dst.Exec("CREATE TABLE transfer_logins (id INTEGER PRIMARY KEY, name VARCHAR(64) NOT NULL)").Error)
					require.NoError(t, dst.Exec("CREATE UNIQUE INDEX uix_transfer_logins_name ON transfer_logins (lower(name))").Error)
				default:
					require.NoError(t, dst.Exec("CREATE TABLE transfer_logins (id INTEGER PRIMARY KEY, name VARCHAR(64) NOT NULL COLLATE NOCASE UNIQUE)").Error)
				}

				t.Cleanup(func() { _ = dst.Exec("DROP TABLE transfer_logins").Error })

				list := Tables{"transfer_logins": &transferTestLogin{}}

				// Without registration, a duplicate key stops the transfer.
				_, err = list.Transfer(src, dst, TransferOptions{BatchSize: 2})
				require.Error(t, err)
				assert.Contains(t, err.Error(), "copy transfer_logins")

				TransferSkipDuplicates["transfer_logins"] = true
				t.Cleanup(func() { delete(TransferSkipDuplicates, "transfer_logins") })

				// Duplicates are skipped in every batch, and the paging continues after them.
				result, err := list.Transfer(src, dst, TransferOptions{BatchSize: 2})
				require.NoError(t, err)
				assert.Equal(t, int64(4), result["transfer_logins"])

				var names []string
				require.NoError(t, dst.Table("transfer_logins").Order("id").Pluck("name", &names).Error)
				assert.Equal(t, []string{"admin", "other", "more", "last"}, names)

				// Other errors still stop it, naming the rows read so far, skipped ones included.
				require.NoError(t, src.Exec("DELETE FROM transfer_logins WHERE id = 5").Error)
				require.NoError(t, src.Exec("UPDATE transfer_logins SET name = NULL WHERE id = 6").Error)
				_, err = list.Transfer(src, dst, TransferOptions{BatchSize: 2})
				require.Error(t, err)
				assert.Contains(t, err.Error(), "(rows 5 to 5)")
			})
		}
	})
	t.Run("Panic", func(t *testing.T) {
		dst := transferTestTarget(t, dsn.DriverSQLite3)
		list := Tables{Camera{}.TableName(): &Camera{}, Lens{}.TableName(): &Lens{}}

		assert.Panics(t, func() {
			_, _ = list.Transfer(Db(), dst, TransferOptions{Progress: func(string, int64) { panic("test") }})
		})

		var cameras int64
		require.NoError(t, dst.Model(&Camera{}).Count(&cameras).Error)
		assert.Zero(t, cameras)
	})
	t.Run("NoPrimaryKey", func(t *testing.T) {
		src, err := OpenDb(dsn.DriverSQLite3, filepath.Join(t.TempDir(), "source.db"))
		require.NoError(t, err)
		t.Cleanup(func() { _ = CloseDb(src) })
		require.NoError(t, src.Exec("CREATE TABLE transfer_nokey (name TEXT)").Error)
		dst := transferTestTarget(t, dsn.DriverSQLite3)
		require.NoError(t, dst.Exec("CREATE TABLE transfer_nokey (name TEXT)").Error)

		_, err = Tables{"transfer_nokey": &transferTestNoKey{}}.Transfer(src, dst, TransferOptions{})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "no primary key")
	})
}

// transferTestNull is a row with a nullable composite primary key.
type transferTestNull struct {
	A string  `gorm:"primaryKey"`
	B *string `gorm:"primaryKey"`
}

// TableName returns the table name of transferTestNull.
func (transferTestNull) TableName() string {
	return "transfer_nulls"
}

// transferTestBytes is a named byte slice that converts itself to a database value.
type transferTestBytes []byte

// Value returns the database value of transferTestBytes.
func (b transferTestBytes) Value() (driver.Value, error) {
	return string(b), nil
}

// transferTestLogin is a row of a log table whose names are unique.
type transferTestLogin struct {
	ID   uint `gorm:"primaryKey"`
	Name *string
}

// TableName returns the table name of transferTestLogin.
func (transferTestLogin) TableName() string {
	return "transfer_logins"
}

// transferTestNoKey is a row without a primary key.
type transferTestNoKey struct {
	Name string
}

// TableName returns the table name of transferTestNoKey.
func (transferTestNoKey) TableName() string {
	return "transfer_nokey"
}

func TestTables_Unknown(t *testing.T) {
	db, err := OpenDb(dsn.DriverSQLite3, filepath.Join(t.TempDir(), "unknown.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = CloseDb(db) })

	t.Run("Empty", func(t *testing.T) {
		unknown, err := Entities.Unknown(db)

		require.NoError(t, err)
		assert.Empty(t, unknown)
	})
	t.Run("Tables", func(t *testing.T) {
		require.NoError(t, db.AutoMigrate(&Lens{}))
		require.NoError(t, db.Exec("CREATE TABLE other (id INTEGER PRIMARY KEY AUTOINCREMENT)").Error)
		require.NoError(t, db.Exec("INSERT INTO other DEFAULT VALUES").Error)
		require.NoError(t, db.Exec("CREATE TABLE subjects_dev1 (id INTEGER)").Error)

		unknown, err := Entities.Unknown(db)

		require.NoError(t, err)
		assert.Equal(t, []string{"other"}, unknown)
	})
	t.Run("TestDatabase", func(t *testing.T) {
		unknown, err := Entities.Unknown(Db())

		require.NoError(t, err)
		assert.NotContains(t, unknown, Photo{}.TableName())
	})
}

func TestTransferValue(t *testing.T) {
	t.Run("EmptyRawMessage", func(t *testing.T) {
		v := transferValue(json.RawMessage{})

		require.IsType(t, []byte{}, v)
		assert.NotNil(t, v.([]byte))
		assert.Empty(t, v.([]byte))
	})
	t.Run("NilRawMessage", func(t *testing.T) {
		v := transferValue(json.RawMessage(nil))

		require.IsType(t, []byte{}, v)
		assert.Nil(t, v.([]byte))
	})
	t.Run("RawMessage", func(t *testing.T) {
		assert.Equal(t, []byte(`{"a":1}`), transferValue(json.RawMessage(`{"a":1}`)))
	})
	t.Run("Valuer", func(t *testing.T) {
		v := transferValue(transferTestBytes{})

		assert.IsType(t, transferTestBytes{}, v)
	})
	t.Run("Other", func(t *testing.T) {
		assert.Equal(t, "zz", transferValue("zz"))
		assert.Equal(t, []byte("x"), transferValue([]byte("x")))
		assert.Nil(t, transferValue(nil))
	})
}

func TestTransferStateOf(t *testing.T) {
	db, err := OpenDb(dsn.DriverSQLite3, filepath.Join(t.TempDir(), "state.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = CloseDb(db) })
	require.NoError(t, db.AutoMigrate(&Lens{}, &PhotoKeyword{}))

	t.Run("Empty", func(t *testing.T) {
		state, err := transferStateOf(db, Lens{}.TableName(), &Lens{})

		require.NoError(t, err)
		assert.Equal(t, transferState{}, state)
	})
	t.Run("Rows", func(t *testing.T) {
		require.NoError(t, db.Create(&Lens{LensSlug: "a", LensName: "A"}).Error)

		state, err := transferStateOf(db, Lens{}.TableName(), &Lens{})

		require.NoError(t, err)
		assert.Equal(t, int64(1), state.Rows)
		assert.NotEmpty(t, state.Updated)
	})
	t.Run("NoUpdatedAt", func(t *testing.T) {
		require.NoError(t, db.Create(&PhotoKeyword{PhotoID: 1, KeywordID: 1}).Error)

		state, err := transferStateOf(db, PhotoKeyword{}.TableName(), &PhotoKeyword{})

		require.NoError(t, err)
		assert.Equal(t, transferState{Rows: 1}, state)
	})
	t.Run("MissingTable", func(t *testing.T) {
		_, err := transferStateOf(db, Camera{}.TableName(), &Camera{})

		assert.Error(t, err)
	})
}

func TestIsDuplicateKeyError(t *testing.T) {
	assert.True(t, isDuplicateKeyError(&mysql.MySQLError{Number: 1062}))
	assert.False(t, isDuplicateKeyError(&mysql.MySQLError{Number: 1213}))
	assert.True(t, isDuplicateKeyError(fmt.Errorf("insert: %w", &pgconn.PgError{Code: "23505"})))
	assert.False(t, isDuplicateKeyError(&pgconn.PgError{Code: "23502"}))
	assert.True(t, isDuplicateKeyError(sqlite3.Error{Code: sqlite3.ErrConstraint, ExtendedCode: sqlite3.ErrConstraintUnique}))
	assert.True(t, isDuplicateKeyError(sqlite3.Error{Code: sqlite3.ErrConstraint, ExtendedCode: sqlite3.ErrConstraintPrimaryKey}))
	assert.False(t, isDuplicateKeyError(sqlite3.Error{Code: sqlite3.ErrConstraint, ExtendedCode: sqlite3.ErrConstraintNotNull}))
	assert.False(t, isDuplicateKeyError(errors.New("duplicate")))
	assert.False(t, isDuplicateKeyError(nil))
}

func TestOtherConnections(t *testing.T) {
	t.Run("SQLite", func(t *testing.T) {
		db, err := OpenDb(dsn.DriverSQLite3, filepath.Join(t.TempDir(), "other.db"))
		require.NoError(t, err)
		t.Cleanup(func() { _ = CloseDb(db) })

		count, err := OtherConnections(db)

		require.NoError(t, err)
		assert.Zero(t, count)
	})
	t.Run("TestDatabase", func(t *testing.T) {
		driver := dsn.ParseDriver(Db().Name())

		if driver == dsn.DriverSQLite3 {
			t.Skip("requires a MariaDB or PostgreSQL test database")
		}

		before, err := OtherConnections(Db())
		require.NoError(t, err)

		// Connections of the same pool are not counted.
		sqlDb, err := Db().DB()
		require.NoError(t, err)

		for i := 0; i < 3; i++ {
			conn, connErr := sqlDb.Conn(context.Background())
			require.NoError(t, connErr)
			t.Cleanup(func() { _ = conn.Close() })
		}

		same, err := OtherConnections(Db())
		require.NoError(t, err)
		assert.Less(t, same, before+3)

		// A connection of another pool to the same database is counted.
		other, err := OpenDb(driver, TestDbDSN(os.Getenv("PHOTOPRISM_TEST_DRIVER"), os.Getenv("PHOTOPRISM_TEST_DSN")))
		require.NoError(t, err)
		t.Cleanup(func() { _ = CloseDb(other) })
		require.NoError(t, other.Exec("SELECT 1").Error)

		after, err := OtherConnections(Db())
		require.NoError(t, err)
		assert.GreaterOrEqual(t, after, same+1)
	})
	t.Run("NotConnected", func(t *testing.T) {
		_, err := OtherConnections(nil)

		assert.Error(t, err)
	})
}

func TestUserNameCollisions(t *testing.T) {
	for _, driver := range []string{dsn.DriverSQLite3, dsn.DriverMySQL, dsn.DriverPostgres} {
		t.Run(driver, func(t *testing.T) {
			db := transferTestTarget(t, driver)

			count, err := UserNameCollisions(db)
			require.NoError(t, err)
			assert.Zero(t, count)

			// Accounts with the same name, names that differ only by case and accents, and accounts without a name.
			for i, name := range []string{"same", "same", "rene", "Ren\u00e9", "", ""} {
				require.NoError(t, db.Exec("INSERT INTO auth_users (id, user_uid, user_name) VALUES (?, ?, ?)", 100+i, rnd.GenerateUID(UserUID), name).Error)
			}

			// Deleted accounts are included, as name lookups for sign-in are unscoped.
			require.NoError(t, db.Exec("UPDATE auth_users SET deleted_at = ? WHERE id = 101", time.Now().UTC()).Error)

			count, err = UserNameCollisions(db)
			require.NoError(t, err)

			if driver == dsn.DriverMySQL {
				assert.Equal(t, int64(4), count)
			} else {
				assert.Equal(t, int64(2), count)
			}
		})
	}
	t.Run("NotConnected", func(t *testing.T) {
		_, err := UserNameCollisions(nil)

		assert.Error(t, err)
	})
}

func TestTransferTxOptions(t *testing.T) {
	sqlite, err := OpenDb(dsn.DriverSQLite3, ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = CloseDb(sqlite) })

	assert.Nil(t, transferTxOptions(sqlite))

	if dsn.ParseDriver(Db().Name()) != dsn.DriverSQLite3 {
		assert.Equal(t, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true}, transferTxOptions(Db()))
	}
}
