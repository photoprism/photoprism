package dsn

import (
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
				Driver:   DriverPostgres,
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
				Driver:   DriverPostgres,
				User:     "alice",
				Password: "s3cr3t",
				Server:   "db.internal:5432",
				Name:     "app",
				Params:   "connect_timeout=5 sslmode=require",
			},
			ok: true,
		},
		{
			name: "QuotedValues",
			in:   `user="alice" password="s ec ret" dbname="app" host=db.internal`,
			want: DSN{
				DSN:      `user="alice" password="s ec ret" dbname="app" host=db.internal`,
				Driver:   DriverPostgres,
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
			want: DSN{DSN: "user=alice host=db.internal"},
			ok:   false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			d := DSN{DSN: tt.in}
			ok := d.parsePostgres()

			assert.Equal(t, tt.in, d.String())

			if ok != tt.ok {
				t.Fatalf("parsePostgres(%q) ok=%v, want %v", tt.in, ok, tt.ok)
			}

			if ok && d != tt.want {
				t.Fatalf("parsePostgres(%q) = %#v, want %#v", tt.in, d, tt.want)
			}
		})
	}
}

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
			want: "postgresql://myuser:password@postgres:5432/testdb?sslmode=disable&TimeZone=UTC&lock_timeout=5000",
		},
		{
			name: "PostgresWithParms",
			in: DSN{
				Driver:   DriverPostgres,
				Name:     "testdb",
				Server:   "postgres:5432",
				User:     "myuser",
				Password: "password",
				Params:   "sslmode=require&TimeZone=UTC&lock_timeout=5000",
			},
			want: "postgresql://myuser:password@postgres:5432/testdb?sslmode=require&TimeZone=UTC&lock_timeout=5000",
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
			want: "postgresql://myuser:password@postgres:5432/testdb?sslmode=disable&TimeZone=UTC&lock_timeout=5000",
		},
		{
			name: "PostgreSQLWithParms",
			in: DSN{
				Driver:   DriverPostgreSQL,
				Name:     "testdb",
				Server:   "postgres:5432",
				User:     "myuser",
				Password: "password",
				Params:   "sslmode=require&TimeZone=UTC&lock_timeout=5000",
			},
			want: "postgresql://myuser:password@postgres:5432/testdb?sslmode=require&TimeZone=UTC&lock_timeout=5000",
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
			want: "postgresql://myuser:spec%5Bchar$%402&@postgres:5432/testdb?sslmode=disable&TimeZone=UTC&lock_timeout=5000",
		},
		{
			name: "PostgreSQLWithParmsEncodedPassword",
			in: DSN{
				Driver:   DriverPostgreSQL,
				Name:     "testdb",
				Server:   "postgres:5432",
				User:     "myuser",
				Password: "spec[char@$2&",
				Params:   "sslmode=require&TimeZone=UTC&lock_timeout=5000",
			},
			want: "postgresql://myuser:spec%5Bchar%40$2&@postgres:5432/testdb?sslmode=require&TimeZone=UTC&lock_timeout=5000",
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

			assert.Equal(t, tt.want, d.ToString())
		})
	}
}

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
				Params:   "sslmode=require&TimeZone=UTC&lock_timeout=5000",
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
				Params:   "sslmode=require&TimeZone=UTC&lock_timeout=5000",
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
