package entity

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/dsn"
)

const (
	// postgresTimeZone is the session time zone of PostgreSQL connections whose DSN does not set one.
	postgresTimeZone = "UTC"
	// postgresMaxVarchar is the largest size of a varchar column.
	postgresMaxVarchar = 10485760
)

// postgresDialector is the PostgreSQL dialector with column types that match those of the other databases.
type postgresDialector struct {
	*postgres.Dialector
}

// Initialize opens the connection pool unless one was provided, and registers the ORM callbacks.
func (d postgresDialector) Initialize(db *gorm.DB) error {
	if d.Conn == nil {
		config, err := pgx.ParseConfig(postgresDSN(d.DSN))

		// The parse error is not returned, as it may quote parts of the password.
		if err != nil {
			return errors.New("invalid postgres connection string, special characters must be percent-encoded")
		}

		if !hasPostgresTimeZone(config.RuntimeParams) {
			config.RuntimeParams["timezone"] = postgresTimeZone
		}

		d.Conn = stdlib.OpenDB(*config, stdlib.OptionAfterConnect(postgresAfterConnect))
	}

	return d.Dialector.Initialize(db)
}

// DataTypeOf returns the column type of a field. Binary strings, the VARBINARY columns of MariaDB, are
// text compared byte by byte, as the driver sends strings as text, which the server would parse as escaped
// bytes. Floats without a precision are binary floating-point numbers, and the MySQL TINYINT is a smallint.
func (d postgresDialector) DataTypeOf(field *schema.Field) string {
	switch {
	case field.DataType == schema.Bytes && field.IndirectFieldType.Kind() == reflect.String:
		if field.Size > 0 && field.Size <= postgresMaxVarchar {
			return fmt.Sprintf(`varchar(%d) COLLATE "C"`, field.Size)
		}

		return `text COLLATE "C"`
	case field.DataType == schema.Float && field.Precision <= 0 && field.Size > 0 && field.Size <= 32:
		return "real"
	case field.DataType == schema.Float && field.Precision <= 0:
		return "double precision"
	case strings.HasPrefix(strings.ToLower(string(field.DataType)), "tinyint"):
		return "smallint"
	default:
		return d.Dialector.DataTypeOf(field)
	}
}

// Migrator returns the schema migrator of the dialector.
func (d postgresDialector) Migrator(db *gorm.DB) gorm.Migrator {
	m := d.Dialector.Migrator(db).(postgres.Migrator)
	m.Dialector = d

	return postgresMigrator{Migrator: m, dialector: *d.Dialector}
}

// postgresMigrator is the PostgreSQL migrator, using the column types of postgresDialector.
type postgresMigrator struct {
	postgres.Migrator
	dialector postgres.Dialector
}

// ColumnTypes returns the column types of a table. They are read with the driver's dialector, which reads
// a row of the table without a prepared statement, as a cached one fails once the table has changed.
func (m postgresMigrator) ColumnTypes(value any) ([]gorm.ColumnType, error) {
	inner := m.Migrator
	inner.Dialector = m.dialector

	return inner.ColumnTypes(value)
}

// postgresAfterConnect reads timestamps of new connections in UTC, as the other databases return them.
func postgresAfterConnect(_ context.Context, conn *pgx.Conn) error {
	conn.TypeMap().RegisterType(&pgtype.Type{
		Name:  "timestamptz",
		OID:   pgtype.TimestamptzOID,
		Codec: &pgtype.TimestamptzCodec{ScanLocation: time.UTC},
	})

	return nil
}

// hasPostgresTimeZone reports whether connection parameters set the session time zone.
func hasPostgresTimeZone(params map[string]string) bool {
	for key := range params {
		if strings.EqualFold(key, "timezone") {
			return true
		}
	}

	return false
}

// postgresDSN returns a connection URI with a lowercase scheme, as the driver reads URIs by their scheme.
func postgresDSN(s string) string {
	if scheme, rest, found := strings.Cut(s, "://"); found && dsn.ParseDriver(scheme) == dsn.DriverPostgres {
		return strings.ToLower(scheme) + "://" + rest
	}

	return s
}

// postgresSequence is a column whose default value is taken from a sequence.
type postgresSequence struct {
	Table  string
	Column string
}

// SyncSequences advances the ID sequences of the registered tables on PostgreSQL past the largest ID in
// their column, as rows inserted with explicit IDs do not advance them. It does not move a sequence back
// unless another session takes a value at the same time, and reports and skips a table that fails.
func SyncSequences(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	} else if dsn.DialectFromDriver(db.Name()) != dsn.DialectPostgreSQL {
		return nil
	}

	tables := make([]string, 0, len(Entities))

	for name := range Entities {
		tables = append(tables, name)
	}

	var sequences []postgresSequence

	if err := db.Raw(`SELECT table_name AS "table", column_name AS "column" FROM information_schema.columns
		WHERE table_schema = CURRENT_SCHEMA() AND table_name IN (?) AND (column_default LIKE 'nextval(%' OR is_identity = 'YES')
		ORDER BY table_name, column_name`, tables).Scan(&sequences).Error; err != nil {
		return err
	}

	var errs []error

	for _, s := range sequences {
		if err := syncSequence(db, s); err != nil {
			errs = append(errs, fmt.Errorf("%s (sync sequence of %s.%s)", err, s.Table, s.Column))
		}
	}

	return errors.Join(errs...)
}

// syncSequence advances the sequence of a column of a registered table past the largest value in the column.
func syncSequence(db *gorm.DB, s postgresSequence) error {
	if _, ok := Entities[s.Table]; !ok {
		return fmt.Errorf("unknown table")
	}

	var seq sql.NullString

	if err := db.Raw("SELECT pg_get_serial_sequence(quote_ident(?), ?)", s.Table, s.Column).Row().Scan(&seq); err != nil {
		return err
	} else if !seq.Valid || seq.String == "" {
		return nil
	}

	var maxID int64

	if err := db.Raw("SELECT COALESCE(MAX(?), 0) FROM ?", clause.Column{Name: s.Column}, clause.Table{Name: s.Table}).Row().Scan(&maxID); err != nil {
		return err
	}

	// The last value is NULL if the sequence was not used since it was created or set.
	var last sql.NullInt64

	if err := db.Raw("SELECT q.last_value FROM pg_sequences q JOIN pg_class c ON c.relname = q.sequencename "+
		"JOIN pg_namespace n ON n.oid = c.relnamespace AND n.nspname = q.schemaname WHERE c.oid = ?::regclass",
		seq.String).Row().Scan(&last); err != nil {
		return err
	} else if last.Valid && last.Int64 >= maxID {
		return nil
	}

	// Taking the next value keeps a sequence that is ahead, also when it was set without being used.
	return db.Exec("SELECT setval(?::regclass, GREATEST(?, nextval(?::regclass)), false)", seq.String, maxID+1, seq.String).Error
}

// ErrPostgresSchemaUnsupported is returned for a PostgreSQL database whose schema cannot be migrated.
var ErrPostgresSchemaUnsupported = errors.New("database was created by an unsupported development version of " +
	"PostgreSQL support, please create a new database and index your originals again")

// CheckPostgresSchema returns ErrPostgresSchemaUnsupported if a PostgreSQL database stores binary strings as
// bytea, as databases created by development versions do, since migrating them would alter the stored values.
func CheckPostgresSchema(db *gorm.DB) error {
	if db == nil || dsn.DialectFromDriver(db.Dialector.Name()) != dsn.DialectPostgreSQL {
		return nil
	}

	var dataType string

	err := db.Raw("SELECT data_type FROM information_schema.columns WHERE table_schema = current_schema() " +
		"AND table_name = 'photos' AND column_name = 'photo_uid'").Row().Scan(&dataType)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil
	case err != nil:
		return fmt.Errorf("failed to check the database schema (%s)", clean.Error(err))
	case dataType == "bytea":
		return ErrPostgresSchemaUnsupported
	}

	return nil
}
