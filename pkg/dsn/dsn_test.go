package dsn

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

//nolint:gosec // G101: DSN parsing tests intentionally use inline credential samples.
func TestDSN_HostAndPort(t *testing.T) {
	tests := []struct {
		name string
		in   string
		host string
		port int
	}{
		{
			name: "MySQLTCP",
			in:   "user:secret@tcp(localhost:3307)/photoprism?parseTime=true",
			host: "localhost",
			port: 3307,
		},
		{
			name: "MySQLIPv6",
			in:   "user:secret@tcp([2001:db8::1]:3307)/photoprism",
			host: "2001:db8::1",
			port: 3307,
		},
		{
			name: "MySQLDefaultPort",
			in:   "user:secret@tcp(mysql.local)/photoprism",
			host: "mysql.local",
			port: 3306,
		},
		{
			name: "PostgresURL",
			in:   "postgres://user:secret@localhost/mydb",
			host: "localhost",
			port: 5432,
		},
		{
			name: "PostgresKeyValue",
			in:   "user=alice password=secret host=/var/run/postgresql port=6432 dbname=app",
			host: "/var/run/postgresql",
			port: 6432,
		},
		{
			name: "PostgresPortOnly",
			in:   "user=alice password=secret port=5433 dbname=app",
			host: "",
			port: 5433,
		},
		{
			name: "SQLite",
			in:   "file:/data/index.db",
			host: "",
			port: 0,
		},
		{
			name: "InvalidPortFallback",
			in:   "user:secret@tcp(localhost:abc)/photoprism",
			host: "localhost",
			port: 3306,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Parse(tt.in)
			assert.Equal(t, tt.host, d.Host())
			assert.Equal(t, tt.port, d.Port())
		})
	}
}

func TestDSN_MaskPassword(t *testing.T) {
	d := Parse("user:secret@tcp(localhost:3306)/db")
	assert.Equal(t, "user:***@tcp(localhost:3306)/db", d.MaskPassword())

	p := Parse("user=alice password=s3cr3t dbname=app")
	assert.Equal(t, "user=alice password=*** dbname=app", p.MaskPassword())

	noPass := Parse("user@tcp(localhost:3306)/db")
	assert.Equal(t, "user@tcp(localhost:3306)/db", noPass.MaskPassword())
}

//nolint:gosec // G101: DSN parsing tests intentionally use inline credential samples.
func TestDSN_ParsePostgres(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want DSN
		ok   bool
	}{
		{
			name: "Basic",
			in:   "user=alice password=s3cr3t dbname=app",
			want: DSN{
				DSN:      "user=alice password=s3cr3t dbname=app",
				Driver:   DriverPostgreSQL,
				User:     "alice",
				Password: "s3cr3t",
				Name:     "app",
			},
			ok: true,
		},
		{
			name: "WithHostPortAndParams",
			in:   "user=alice password=s3cr3t dbname=app host=db.internal port=5432 connect_timeout=5 sslmode=require",
			want: DSN{
				DSN:      "user=alice password=s3cr3t dbname=app host=db.internal port=5432 connect_timeout=5 sslmode=require",
				Driver:   DriverPostgreSQL,
				User:     "alice",
				Password: "s3cr3t",
				Server:   "db.internal:5432",
				Name:     "app",
				Params:   "connect_timeout=5&sslmode=require",
			},
			ok: true,
		},
		{
			name: "QuotedValues",
			in:   `user="alice" password="s ec ret" dbname="app" host=db.internal`,
			want: DSN{
				DSN:      `user="alice" password="s ec ret" dbname="app" host=db.internal`,
				Driver:   DriverPostgreSQL,
				User:     "alice",
				Password: "s ec ret",
				Server:   "db.internal",
				Name:     "app",
			},
			ok: true,
		},
		{
			name: "MissingDatabase",
			in:   "user=alice host=db.internal",
			want: DSN{DSN: "user=alice host=db.internal"}, // Parsing should abort as dbname is missing.
			ok:   false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			d := DSN{DSN: tt.in}
			ok := d.parsePostgres()

			assert.Equal(t, tt.in, d.String())
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.want, d)
		})
	}
}

//nolint:gosec // G101: DSN parsing tests intentionally use inline credential samples.
func TestDSN_ToString(t *testing.T) {
	cases := []struct {
		name string
		in   DSN
		want string
	}{
		{
			name: "NoDriver",
			in: DSN{
				Name:   "test.db",
				Server: "storage/unittest",
			},
			want: "storage/unittest/test.db?_busy_timeout=5000&_foreign_keys=on",
		},
		{
			name: "NoDriverWithParms",
			in: DSN{
				Name:   "test.db",
				Server: "storage/unittest",
				Params: "_busy_timeout=15000&_foreign_keys=off",
			},
			want: "storage/unittest/test.db?_busy_timeout=15000&_foreign_keys=off",
		},
		{
			name: "SQLite",
			in: DSN{
				Driver: DriverSQLite3,
				Name:   "test.db",
				Server: "storage/unittest",
			},
			want: "storage/unittest/test.db?_busy_timeout=5000&_foreign_keys=on",
		},
		{
			name: "SQLiteWithParms",
			in: DSN{
				Driver: DriverSQLite3,
				Name:   "test.db",
				Server: "storage/unittest",
				Params: "_busy_timeout=15000&_foreign_keys=off",
			},
			want: "storage/unittest/test.db?_busy_timeout=15000&_foreign_keys=off",
		},
		{
			name: "SQLitefile",
			in: DSN{
				Driver: "sqlitefile",
				Name:   "test.db",
				Server: "storage/unittest",
			},
			want: "storage/unittest/test.db?_busy_timeout=5000&_foreign_keys=on",
		},
		{
			name: "SQLitefileWithParms",
			in: DSN{
				Driver: "sqlitefile",
				Name:   "test.db",
				Server: "storage/unittest",
				Params: "_busy_timeout=15000&_foreign_keys=off",
			},
			want: "storage/unittest/test.db?_busy_timeout=15000&_foreign_keys=off",
		},
		{
			name: "Postgres",
			in: DSN{
				Driver:   DriverPostgres,
				Name:     "testdb",
				Server:   "postgres:5432",
				User:     "myuser",
				Password: "password",
			},
			want: "postgresql://myuser:password@postgres:5432/testdb?sslmode=disable&TimeZone=UTC",
		},
		{
			name: "PostgresWithParms",
			in: DSN{
				Driver:   DriverPostgres,
				Name:     "testdb",
				Server:   "postgres:5432",
				User:     "myuser",
				Password: "password",
				Params:   "sslmode=require TimeZone=UTC",
			},
			want: "postgresql://myuser:password@postgres:5432/testdb?sslmode=require&TimeZone=UTC",
		},
		{
			name: "PostgreSQL",
			in: DSN{
				Driver:   DriverPostgreSQL,
				Name:     "testdb",
				Server:   "postgres:5432",
				User:     "myuser",
				Password: "password",
			},
			want: "postgresql://myuser:password@postgres:5432/testdb?sslmode=disable&TimeZone=UTC",
		},
		{
			name: "PostgreSQLWithParms",
			in: DSN{
				Driver:   DriverPostgreSQL,
				Name:     "testdb",
				Server:   "postgres:5432",
				User:     "myuser",
				Password: "password",
				Params:   "sslmode=require&TimeZone=UTC&connect_timeout=5000",
			},
			want: "postgresql://myuser:password@postgres:5432/testdb?sslmode=require&TimeZone=UTC&connect_timeout=5000",
		},
		{
			name: "PostgreSQLEncodedPassword",
			in: DSN{
				Driver:   DriverPostgreSQL,
				Name:     "testdb",
				Server:   "postgres:5432",
				User:     "myuser",
				Password: "spec[char$@2&",
			},
			want: "postgresql://myuser:spec%5Bchar$%402&@postgres:5432/testdb?sslmode=disable&TimeZone=UTC",
		},
		{
			name: "PostgreSQLWithParmsEncodedPassword",
			in: DSN{
				Driver:   DriverPostgreSQL,
				Name:     "testdb",
				Server:   "postgres:5432",
				User:     "myuser",
				Password: "spec[char@$2&",
				Params:   "sslmode=require&TimeZone=UTC&connect_timeout=5000",
			},
			want: "postgresql://myuser:spec%5Bchar%40$2&@postgres:5432/testdb?sslmode=require&TimeZone=UTC&connect_timeout=5000",
		},
		{
			name: "MariaDB",
			in: DSN{
				Driver:   DriverMariaDB,
				Name:     "testdb",
				Server:   "mariadb:4001",
				User:     "myuser",
				Password: "password",
			},
			want: "myuser:password@mariadb:4001/testdb?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true",
		},
		{
			name: "MariaDBNet",
			in: DSN{
				Driver:   DriverMariaDB,
				Name:     "testdb",
				Server:   "mariadb:4001",
				User:     "myuser",
				Password: "password",
				Net:      "tcp",
			},
			want: "myuser:password@tcp(mariadb:4001)/testdb?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true",
		},
		{
			name: "MariaDBWithParms",
			in: DSN{
				Driver:   DriverMariaDB,
				Name:     "testdb",
				Server:   "mariadb:4001",
				User:     "myuser",
				Password: "password",
				Params:   "charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true&useSSL=false",
			},
			want: "myuser:password@mariadb:4001/testdb?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true&useSSL=false",
		},
		{
			name: "MySQL",
			in: DSN{
				Driver:   DriverMySQL,
				Name:     "testdb",
				Server:   "mariadb:4001",
				User:     "myuser",
				Password: "password",
			},
			want: "myuser:password@mariadb:4001/testdb?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true",
		},
		{
			name: "MySQLWithParms",
			in: DSN{
				Driver:   DriverMySQL,
				Name:     "testdb",
				Server:   "mariadb:4001",
				User:     "myuser",
				Password: "password",
				Params:   "charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true&useSSL=false",
			},
			want: "myuser:password@mariadb:4001/testdb?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true&useSSL=false",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.in

			if d.Driver == DriverPostgreSQL || d.Driver == DriverPostgres {
				// Postgres uses a URI, which means that the Query part is not ordered.  Causing string assert.Equal to fail (sometimes).
				expected, eerr := url.Parse(tt.want)
				actual, aerr := url.Parse(d.ToString())
				assert.Equal(t, eerr, aerr)
				exQ := expected.Query()
				acQ := actual.Query()
				expected.RawQuery = ""
				actual.RawQuery = ""
				assert.Equal(t, expected, actual)
				exJ, _ := json.Marshal(exQ)
				acJ, _ := json.Marshal(acQ)
				assert.JSONEq(t, string(exJ), string(acJ))
			} else {
				assert.Equal(t, tt.want, d.ToString())
			}
		})
	}
}

//nolint:gosec // G101: DSN parsing tests intentionally use inline credential samples.
func TestDSN_ForPSQL(t *testing.T) {
	cases := []struct {
		name string
		in   DSN
		want string
	}{
		{
			name: "Postgres",
			in: DSN{
				Driver:   DriverPostgres,
				Name:     "testdb",
				Server:   "postgres:5432",
				User:     "myuser",
				Password: "password",
			},
			want: "postgresql://myuser:password@postgres:5432/testdb",
		},
		{
			name: "PostgresWithParms",
			in: DSN{
				Driver:   DriverPostgres,
				Name:     "testdb",
				Server:   "postgres:5432",
				User:     "myuser",
				Password: "password",
				Params:   "sslmode=require&TimeZone=UTC&connect_timeout=5000",
			},
			want: "postgresql://myuser:password@postgres:5432/testdb",
		},
		{
			name: "PostgreSQL",
			in: DSN{
				Driver:   DriverPostgreSQL,
				Name:     "testdb",
				Server:   "postgres:5432",
				User:     "myuser",
				Password: "password",
			},
			want: "postgresql://myuser:password@postgres:5432/testdb",
		},
		{
			name: "PostgreSQLWithParms",
			in: DSN{
				Driver:   DriverPostgreSQL,
				Name:     "testdb",
				Server:   "postgres:5432",
				User:     "myuser",
				Password: "password",
				Params:   "sslmode=require&TimeZone=UTC&connect_timeout=5000",
			},
			want: "postgresql://myuser:password@postgres:5432/testdb",
		},
		{
			name: "MariaDB",
			in: DSN{
				Driver:   DriverMariaDB,
				Name:     "testdb",
				Server:   "mariadb:4001",
				User:     "myuser",
				Password: "password",
			},
			want: "postgresql://myuser:password@mariadb:4001/testdb",
		},
		{
			name: "PostgreSQLEncodedPassword",
			in: DSN{
				Driver:   DriverPostgreSQL,
				Name:     "testdb",
				Server:   "postgres:5432",
				User:     "myuser",
				Password: "spec[char$@2&",
			},
			want: "postgresql://myuser:spec%5Bchar$%402&@postgres:5432/testdb",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.in

			assert.Equal(t, tt.want, d.ForPSQL())
		})
	}
}

func TestPostgresEncodeParams(t *testing.T) {
	cases := []struct {
		name string
		in   DSN
		want string
	}{
		{
			name: "SingleParam",
			in: DSN{
				Params: "_busy_timeout=5000",
			},
			want: "_busy_timeout=5000",
		},
		{
			name: "TwoParams",
			in: DSN{
				Params: "_busy_timeout=15000 _foreign_keys=off",
			},
			want: "_busy_timeout=15000&_foreign_keys=off",
		},
		{
			name: "ThreeParams",
			in: DSN{
				Params: "sslmode=require TimeZone=UTC connect_timeout=5000",
			},
			want: "sslmode=require&TimeZone=UTC&connect_timeout=5000",
		},
		{
			name: "AlreadyEncoded",
			in: DSN{
				Params: "_busy_timeout=15000&_foreign_keys=off",
			},
			want: "_busy_timeout=15000&_foreign_keys=off",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.in
			postgresEncodeParams(&d)

			exQ, eerr := url.ParseQuery(tt.want)
			acQ, aerr := url.ParseQuery(d.Params)
			exJ, _ := json.Marshal(exQ)
			acJ, _ := json.Marshal(acQ)
			assert.Nil(t, eerr)
			assert.Equal(t, eerr, aerr)
			assert.JSONEq(t, string(exJ), string(acJ))
		})
	}
}

func TestPostgresDequote(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "Quoted",
			in:   "'quotedstring'",
			want: "quotedstring",
		},
		{
			name: "QuotedEscapedBackSlash",
			in:   `'quoted\\backslash'`,
			want: `quoted\backslash`,
		},
		{
			name: "QuotedEscapedQuote",
			in:   `'quoted\'quote'`,
			want: `quoted'quote`,
		},
		{
			name: "LeadingQuote",
			in:   `'leadingquote`,
			want: `'leadingquote`,
		},
		{
			name: "TrailingQuote",
			in:   `trailingquote'`,
			want: `trailingquote'`,
		},
		{
			name: "EscapedBackSlash",
			in:   `escaped\\backslash`,
			want: `escaped\backslash`,
		},
		{
			name: "EscapedQuote",
			in:   `escaped\'quote`,
			want: `escaped'quote`,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, postgresDequote(tt.in))
		})
	}
}
