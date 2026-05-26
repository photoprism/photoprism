package dsn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//nolint:gosec // G101: DSN parsing tests intentionally use inline credential samples.
func TestParse(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want DSN
	}{
		{
			name: "ClassicTCP",
			in:   "user:secret@tcp(localhost:3306)/photoprism?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true",
			want: DSN{
				DSN:      "user:secret@tcp(localhost:3306)/photoprism?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true",
				Driver:   DriverMySQL,
				User:     "user",
				Password: "secret",
				Net:      "tcp",
				Server:   "localhost:3306",
				Name:     "photoprism",
				Params:   "charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true",
			},
		},
		{
			name: "URIStyle",
			in:   "mysql://user:secret@localhost:3306/photoprism?parseTime=true",
			want: DSN{
				DSN:      "mysql://user:secret@localhost:3306/photoprism?parseTime=true",
				Driver:   DriverMySQL,
				User:     "user",
				Password: "secret",
				Server:   "localhost:3306",
				Name:     "photoprism",
				Params:   "parseTime=true",
			},
		},
		{
			name: "UnixSocket",
			in:   "user:secret@unix(/var/run/mysql.sock)/photoprism",
			want: DSN{
				DSN:      "user:secret@unix(/var/run/mysql.sock)/photoprism",
				Driver:   DriverMySQL,
				User:     "user",
				Password: "secret",
				Net:      "unix",
				Server:   "/var/run/mysql.sock",
				Name:     "photoprism",
			},
		},
		{
			name: "FileDSN",
			in:   "file:/data/index.db?_busy_timeout=5000",
			want: DSN{
				DSN:    "file:/data/index.db?_busy_timeout=5000",
				Driver: DriverSQLite3,
				Server: "file:/data",
				Name:   "index.db",
				Params: "_busy_timeout=5000",
			},
		},
		{
			name: "SQLite",
			in:   "/index.db?_busy_timeout=5000",
			want: DSN{
				DSN:    "/index.db?_busy_timeout=5000",
				Driver: DriverSQLite3,
				Server: "",
				Name:   "index.db",
				Params: "_busy_timeout=5000",
			},
		},
		{
			name: "PostgresKeyValue",
			in:   "user=alice password=s3cr3t dbname=app host=db.internal port=5432 connect_timeout=5 sslmode=require",
			want: DSN{
				DSN:      "user=alice password=s3cr3t dbname=app host=db.internal port=5432 connect_timeout=5 sslmode=require",
				Driver:   DriverPostgres,
				User:     "alice",
				Password: "s3cr3t",
				Server:   "db.internal:5432",
				Name:     "app",
				Params:   "connect_timeout=5 sslmode=require",
			},
		},
		{
			name: "EmptyInput",
			in:   "",
			want: DSN{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Parse(tt.in)
			assert.Equal(t, tt.in, got.String())
			if got != tt.want {
				t.Fatalf("Parse(%q) = %#v, want %#v", tt.in, got, tt.want)
			}
		})
	}
}

// TestParse_DriverDetection exercises detectDriver via Parse: alias unification
// with ParseDriver and the preserve-unknown contract.
func TestParse_DriverDetection(t *testing.T) {
	t.Run("PostgresqlAliasNormalizes", func(t *testing.T) {
		got := Parse("postgresql://alice:s3cr3t@db.local:5432/app")
		assert.Equal(t, DriverPostgres, got.Driver)
	})
	t.Run("PostgresqlAliasCaseInsensitive", func(t *testing.T) {
		got := Parse("POSTGRESQL://alice:s3cr3t@db.local:5432/app")
		assert.Equal(t, DriverPostgres, got.Driver)
	})
	t.Run("MariaDBAliasCollapsesToMySQL", func(t *testing.T) {
		got := Parse("mariadb://user:secret@db.local:3306/app")
		assert.Equal(t, DriverMySQL, got.Driver)
	})
	t.Run("SqliteAliasNormalizes", func(t *testing.T) {
		got := Parse("sqlite:///data/index.db")
		assert.Equal(t, DriverSQLite3, got.Driver)
	})
	t.Run("UnknownDriverPreserved", func(t *testing.T) {
		// Unknown explicit drivers are kept instead of falling through to heuristics.
		got := Parse("oracle://user:secret@db.local:1521/app")
		assert.Equal(t, "oracle", got.Driver)
	})
	t.Run("UnknownDriverPreservedLowercased", func(t *testing.T) {
		got := Parse("Snowflake://user:secret@account.region/db")
		assert.Equal(t, "snowflake", got.Driver)
	})
}
