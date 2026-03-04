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
			name: "PostgreSQLURI1",
			in:   "postgresql://john:pass@postgres:5432/my_db?TimeZone=UTC&connect_timeout=15&lock_timeout=5000&sslmode=disable",
			want: DSN{
				DSN:      "postgresql://john:pass@postgres:5432/my_db?TimeZone=UTC&connect_timeout=15&lock_timeout=5000&sslmode=disable",
				Driver:   DriverPostgres,
				User:     "john",
				Password: "pass",
				Server:   "postgres:5432",
				Name:     "my_db",
				Params:   "TimeZone=UTC&connect_timeout=15&lock_timeout=5000&sslmode=disable",
			},
		},
		{
			name: "PostgreSQLURI2",
			in:   "postgres://john:pass@postgres:5432/my_db?TimeZone=UTC&connect_timeout=15&lock_timeout=5000&sslmode=disable",
			want: DSN{
				DSN:      "postgres://john:pass@postgres:5432/my_db?TimeZone=UTC&connect_timeout=15&lock_timeout=5000&sslmode=disable",
				Driver:   DriverPostgres,
				User:     "john",
				Password: "pass",
				Server:   "postgres:5432",
				Name:     "my_db",
				Params:   "TimeZone=UTC&connect_timeout=15&lock_timeout=5000&sslmode=disable",
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
