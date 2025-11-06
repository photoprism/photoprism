package dsn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
