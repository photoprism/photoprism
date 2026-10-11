package entity

import (
	"encoding/json"
	"errors"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"github.com/photoprism/photoprism/pkg/dsn"
)

// postgresTypeName is a named string type for testing PostgreSQL column types.
type postgresTypeName string

// postgresTypeTest is a model for testing PostgreSQL column types.
type postgresTypeTest struct {
	ID      uint             `gorm:"primaryKey"`
	Name    string           `gorm:"size:160"`
	Slug    string           `gorm:"type:bytes;size:42"`
	Index   *string          `gorm:"type:bytes;size:64"`
	Kind    postgresTypeName `gorm:"type:bytes;size:16"`
	Note    string           `gorm:"type:bytes;size:20000000"`
	Data    json.RawMessage  `gorm:"type:bytes;size:4096"`
	Lat     float64
	Score   float32 `gorm:"size:64"`
	Ratio   float32
	Amount  float64 `gorm:"precision:10;scale:2"`
	Retry   uint8   `gorm:"type:TINYINT(3);size:3;default:0"`
	Count   int     `gorm:"size:32"`
	Enabled bool
}

// TableName returns the entity table name.
func (postgresTypeTest) TableName() string {
	return "type_tests"
}

func TestPostgresDialector_DataTypeOf(t *testing.T) {
	s, err := schema.Parse(&postgresTypeTest{}, &sync.Map{}, schema.NamingStrategy{})
	require.NoError(t, err)

	d := postgresDialector{Dialector: postgres.New(postgres.Config{}).(*postgres.Dialector)}

	assert.Equal(t, "varchar(160)", d.DataTypeOf(s.LookUpField("Name")))
	assert.Equal(t, `varchar(42) COLLATE "C"`, d.DataTypeOf(s.LookUpField("Slug")))
	assert.Equal(t, `varchar(64) COLLATE "C"`, d.DataTypeOf(s.LookUpField("Index")))
	assert.Equal(t, `varchar(16) COLLATE "C"`, d.DataTypeOf(s.LookUpField("Kind")))
	assert.Equal(t, `text COLLATE "C"`, d.DataTypeOf(s.LookUpField("Note")))
	assert.Equal(t, "bytea", d.DataTypeOf(s.LookUpField("Data")))
	assert.Equal(t, "double precision", d.DataTypeOf(s.LookUpField("Lat")))
	assert.Equal(t, "double precision", d.DataTypeOf(s.LookUpField("Score")))
	assert.Equal(t, "real", d.DataTypeOf(s.LookUpField("Ratio")))
	assert.Equal(t, "numeric(10, 2)", d.DataTypeOf(s.LookUpField("Amount")))
	assert.Equal(t, "smallint", d.DataTypeOf(s.LookUpField("Retry")))
	assert.Equal(t, "integer", d.DataTypeOf(s.LookUpField("Count")))
	assert.Equal(t, "boolean", d.DataTypeOf(s.LookUpField("Enabled")))
}

func TestPostgresDialector_Migrator(t *testing.T) {
	t.Run("Wrapper", func(t *testing.T) {
		d := postgresDialector{Dialector: postgres.New(postgres.Config{}).(*postgres.Dialector)}
		m, ok := d.Migrator(nil).(postgresMigrator)
		require.True(t, ok)
		assert.Equal(t, d, m.Migrator.Dialector)
		assert.Equal(t, *d.Dialector, m.dialector)
	})
	t.Run("ChangedTable", func(t *testing.T) {
		db := openPostgresTestDb(t, nil)
		pool, err := db.DB()
		require.NoError(t, err)
		pool.SetMaxOpenConns(1)

		require.NoError(t, db.Exec("DROP TABLE IF EXISTS migrator_tests").Error)
		require.NoError(t, db.Exec("CREATE TABLE migrator_tests (id serial PRIMARY KEY)").Error)
		t.Cleanup(func() { _ = db.Exec("DROP TABLE IF EXISTS migrator_tests").Error })

		// Column types are read again after the table changed on the same connection.
		columns, err := db.Migrator().ColumnTypes("migrator_tests")
		require.NoError(t, err)
		assert.Len(t, columns, 1)
		require.NoError(t, db.Exec("ALTER TABLE migrator_tests ADD COLUMN name varchar(16)").Error)
		columns, err = db.Migrator().ColumnTypes("migrator_tests")
		require.NoError(t, err)
		assert.Len(t, columns, 2)
	})
}

// openPostgresTestDb opens a new connection to the PostgreSQL test database, with the DSN changed by update
// if it is not nil, and skips the test on other databases.
func openPostgresTestDb(t *testing.T, update func(u *url.URL)) *gorm.DB {
	t.Helper()

	if DbDialect() != dsn.DialectPostgreSQL {
		t.Skip("requires the PostgreSQL test database")
	}

	conn, ok := dbConn.(*DbConn)
	require.True(t, ok, "test database provider")

	u, err := url.Parse(conn.Dsn)
	require.NoError(t, err)

	if update != nil {
		update(u)
	}

	db, err := OpenDb(conn.Driver, u.String())
	require.NoError(t, err)
	t.Cleanup(func() { _ = CloseDb(db) })

	return db
}

func TestPostgresDialector_Initialize(t *testing.T) {
	t.Run("InvalidDSN", func(t *testing.T) {
		_, err := OpenDb(dsn.DriverPostgres, "postgres://u:p@localhost:bad/db")
		assert.Error(t, err)
	})
	t.Run("ParseErrorWithoutPassword", func(t *testing.T) {
		_, err := OpenDb(dsn.DriverPostgres, "POSTGRES://u:SEC@RET3@h:bad/db")
		require.Error(t, err)
		assert.NotContains(t, err.Error(), "SEC")
		assert.NotContains(t, err.Error(), "RET3")
	})
	t.Run("DefaultTimeZone", func(t *testing.T) {
		if DbDialect() != dsn.DialectPostgreSQL {
			t.Skip("requires the PostgreSQL test database")
		}

		// New sessions get another time zone unless the connection sets one.
		var name string
		require.NoError(t, Db().Raw("SELECT current_database()").Row().Scan(&name))
		require.NoError(t, Db().Exec("ALTER DATABASE ? SET timezone = 'Asia/Tokyo'", clause.Table{Name: name}).Error)
		t.Cleanup(func() { _ = Db().Exec("ALTER DATABASE ? RESET timezone", clause.Table{Name: name}).Error })

		db := openPostgresTestDb(t, func(u *url.URL) {
			q := u.Query()
			q.Del("TimeZone")
			u.RawQuery = q.Encode()
		})

		var zone string
		require.NoError(t, db.Raw("SHOW TimeZone").Row().Scan(&zone))
		assert.Equal(t, postgresTimeZone, zone)
	})
	t.Run("TimeZoneParam", func(t *testing.T) {
		db := openPostgresTestDb(t, func(u *url.URL) {
			q := u.Query()
			q.Set("TimeZone", "Europe/Berlin")
			u.RawQuery = q.Encode()
		})

		var zone string
		require.NoError(t, db.Raw("SHOW TimeZone").Row().Scan(&zone))
		assert.Equal(t, "Europe/Berlin", zone)

		// Timestamps are read in UTC whatever the time zone of the session.
		var ts time.Time
		require.NoError(t, db.Raw("SELECT TIMESTAMPTZ '2026-10-11 12:00:00+00'").Row().Scan(&ts))
		assert.Equal(t, time.UTC, ts.Location())
		assert.Equal(t, time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC), ts)
	})
}

func TestPostgresDSN(t *testing.T) {
	assert.Equal(t, "postgres://u:p@h/db", postgresDSN("POSTGRES://u:p@h/db"))
	assert.Equal(t, "postgresql://u:p@h/db", postgresDSN("PostgreSQL://u:p@h/db"))
	assert.Equal(t, "host=h user=u", postgresDSN("host=h user=u"))
	assert.Equal(t, "mysql://u:p@h/db", postgresDSN("mysql://u:p@h/db"))
}

func TestHasPostgresTimeZone(t *testing.T) {
	assert.True(t, hasPostgresTimeZone(map[string]string{"TimeZone": "UTC"}))
	assert.True(t, hasPostgresTimeZone(map[string]string{"timezone": "Europe/Berlin"}))
	assert.False(t, hasPostgresTimeZone(map[string]string{"application_name": "photoprism"}))
	assert.False(t, hasPostgresTimeZone(nil))
}

// sequenceTest is a model with a generated ID for testing sequences.
type sequenceTest struct {
	ID   int    `gorm:"primaryKey"`
	Name string `gorm:"size:16"`
}

// TableName returns the entity table name.
func (sequenceTest) TableName() string {
	return "sequence_tests"
}

func TestSyncSequences(t *testing.T) {
	t.Run("NilDb", func(t *testing.T) {
		assert.Error(t, SyncSequences(nil))
	})
	t.Run("OtherDialect", func(t *testing.T) {
		assert.NoError(t, SyncSequences(openTestDb(t)))
	})
	t.Run("ExplicitIDs", func(t *testing.T) {
		if DbDialect() != dsn.DialectPostgreSQL {
			t.Skip("requires the PostgreSQL test database")
		}

		db := Db()
		name := sequenceTest{}.TableName()
		require.NoError(t, db.Migrator().DropTable(name))
		require.NoError(t, db.AutoMigrate(&sequenceTest{}))
		Entities[name] = &sequenceTest{}
		t.Cleanup(func() {
			delete(Entities, name)
			_ = db.Migrator().DropTable(name)
		})

		// Rows with explicit IDs leave the sequence at its start.
		require.NoError(t, db.Exec("INSERT INTO sequence_tests (id, name) VALUES (5, 'a'), (-1, 'b')").Error)
		require.NoError(t, SyncSequences(db))

		var id int
		require.NoError(t, db.Raw("INSERT INTO sequence_tests (name) VALUES ('c') RETURNING id").Row().Scan(&id))
		assert.Equal(t, 6, id)

		// A sequence ahead of the largest ID is not moved back.
		require.NoError(t, db.Exec("DELETE FROM sequence_tests WHERE id = 6").Error)
		require.NoError(t, SyncSequences(db))
		require.NoError(t, db.Raw("INSERT INTO sequence_tests (name) VALUES ('d') RETURNING id").Row().Scan(&id))
		assert.Equal(t, 7, id)

		// The same applies to a sequence that was set but not used yet.
		require.NoError(t, db.Exec("SELECT setval(pg_get_serial_sequence('sequence_tests', 'id'), 20, false)").Error)
		require.NoError(t, SyncSequences(db))
		require.NoError(t, db.Raw("INSERT INTO sequence_tests (name) VALUES ('e') RETURNING id").Row().Scan(&id))
		assert.Equal(t, 20, id)

		// Tables that are not registered are not changed.
		delete(Entities, name)
		require.NoError(t, db.Exec("INSERT INTO sequence_tests (id, name) VALUES (100, 'f')").Error)
		require.NoError(t, SyncSequences(db))
		require.NoError(t, db.Raw("INSERT INTO sequence_tests (name) VALUES ('g') RETURNING id").Row().Scan(&id))
		assert.Equal(t, 21, id)
	})
}

func TestSyncSequence(t *testing.T) {
	t.Run("UnknownTable", func(t *testing.T) {
		assert.Error(t, syncSequence(Db(), postgresSequence{Table: "no_such_table", Column: "id"}))
	})
	t.Run("NoSequence", func(t *testing.T) {
		if DbDialect() != dsn.DialectPostgreSQL {
			t.Skip("requires the PostgreSQL test database")
		}

		assert.NoError(t, syncSequence(Db(), postgresSequence{Table: Photo{}.TableName(), Column: "photo_uid"}))
	})
}

func TestCheckPostgresSchema(t *testing.T) {
	t.Run("NilDb", func(t *testing.T) {
		assert.NoError(t, CheckPostgresSchema(nil))
	})
	t.Run("OtherDialect", func(t *testing.T) {
		assert.NoError(t, CheckPostgresSchema(openTestDb(t)))
	})
	t.Run("PostgreSQL", func(t *testing.T) {
		if DbDialect() != dsn.DialectPostgreSQL {
			t.Skip("requires the PostgreSQL test database")
		}

		errRollback := errors.New("rollback")

		// check runs CheckPostgresSchema with the specified schema as the current one.
		check := func(t *testing.T, schema string) (result error) {
			t.Helper()

			err := Db().Transaction(func(tx *gorm.DB) error {
				require.NoError(t, tx.Exec("SET LOCAL search_path TO "+schema).Error)
				result = CheckPostgresSchema(tx)
				return errRollback
			})

			require.ErrorIs(t, err, errRollback)

			return result
		}

		// The current schema of the test database is supported.
		assert.NoError(t, CheckPostgresSchema(Db()))

		schema := "zz_schema_check"
		require.NoError(t, Db().Exec("DROP SCHEMA IF EXISTS "+schema+" CASCADE").Error)
		require.NoError(t, Db().Exec("CREATE SCHEMA "+schema).Error)
		t.Cleanup(func() { _ = Db().Exec("DROP SCHEMA IF EXISTS " + schema + " CASCADE").Error })

		// A database without a photos table is new, and no current schema has none.
		assert.NoError(t, check(t, schema))
		assert.NoError(t, check(t, "''"))

		require.NoError(t, Db().Exec("CREATE TABLE "+schema+".photos (id SERIAL, photo_uid varchar(42) COLLATE \"C\")").Error)
		assert.NoError(t, check(t, schema))

		require.NoError(t, Db().Exec("ALTER TABLE "+schema+".photos ALTER COLUMN photo_uid TYPE bytea USING photo_uid::bytea").Error)
		assert.ErrorIs(t, check(t, schema), ErrPostgresSchemaUnsupported)
	})
}
