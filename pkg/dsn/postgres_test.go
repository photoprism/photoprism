package dsn

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDSN_PostgreSQL(t *testing.T) {
	t.Run("KeyValueParams", func(t *testing.T) {
		d := DSN{User: "photoprism", Password: "secret", Server: "localhost:5432", Name: "photoprism", Params: "sslmode=disable TimeZone=UTC connect_timeout=10"}
		assert.Equal(t, "postgres://photoprism:secret@localhost:5432/photoprism?TimeZone=UTC&connect_timeout=10&sslmode=disable", d.PostgreSQL())
	})
	t.Run("QueryParams", func(t *testing.T) {
		d := DSN{User: "photoprism", Password: "secret", Server: "db:5433", Name: "photoprism", Params: "sslmode=require&connect_timeout=5"}
		assert.Equal(t, "postgres://photoprism:secret@db:5433/photoprism?connect_timeout=5&sslmode=require", d.PostgreSQL())
	})
	t.Run("SpecialCharacters", func(t *testing.T) {
		d := DSN{User: "photo prism", Password: "p@ss:w/rd?#& '", Server: "localhost:5432", Name: "my db"} //nolint:gosec // G101: test fixture, not a credential.
		u, err := url.Parse(d.PostgreSQL())
		require.NoError(t, err)

		password, _ := u.User.Password()
		assert.Equal(t, "photo prism", u.User.Username())
		assert.Equal(t, "p@ss:w/rd?#& '", password)
		assert.Equal(t, "/my db", u.Path)
		assert.Equal(t, "localhost:5432", u.Host)
	})
	t.Run("UnixSocket", func(t *testing.T) {
		d := DSN{User: "photoprism", Password: "secret", Server: "/var/run/postgresql", Name: "photoprism", Params: "sslmode=disable"}
		u, err := url.Parse(d.PostgreSQL())
		require.NoError(t, err)

		assert.Equal(t, "", u.Host)
		assert.Equal(t, "/var/run/postgresql", u.Query().Get("host"))
		assert.Equal(t, "5432", u.Query().Get("port"))
		assert.Equal(t, "disable", u.Query().Get("sslmode"))
	})
	t.Run("UnixSocketPort", func(t *testing.T) {
		d := DSN{User: "photoprism", Server: "/run/pg", Name: "photoprism", Params: "port=5433"}
		u, err := url.Parse(d.PostgreSQL())
		require.NoError(t, err)

		assert.Equal(t, "5433", u.Query().Get("port"))
	})
	t.Run("NameWithSlashes", func(t *testing.T) {
		d := DSN{User: "u", Password: "s3cret", Server: "h:5432", Name: "x://y/z"}
		s := d.PostgreSQL()

		assert.Equal(t, "postgres://u:s3cret@h:5432/x:%2F%2Fy%2Fz", s)
		assert.NotContains(t, Mask(s), "s3cret")
	})
	t.Run("SocketPathWithSpace", func(t *testing.T) {
		d := DSN{User: "u", Server: "/run/my pg", Name: "x"}
		assert.Contains(t, d.PostgreSQL(), "host=%2Frun%2Fmy%20pg")
	})
	t.Run("InvalidQuery", func(t *testing.T) {
		d := DSN{User: "u", Server: "h", Name: "x", Params: "sslmode=require x=%zz"}
		u, err := url.Parse(d.PostgreSQL())
		require.NoError(t, err)

		// Parameters that are not a valid query are read as key=value pairs, so none is lost.
		assert.Equal(t, "require", u.Query().Get("sslmode"))
	})
	t.Run("NoParams", func(t *testing.T) {
		d := DSN{User: "photoprism", Password: "secret", Server: "localhost", Name: "photoprism"}
		assert.Equal(t, "postgres://photoprism:secret@localhost/photoprism", d.PostgreSQL())
	})
	t.Run("RoundTrip", func(t *testing.T) {
		d := DSN{User: "photoprism", Password: "secret", Server: "localhost:5432", Name: "photoprism", Params: "sslmode=disable"}
		parsed := Parse(d.PostgreSQL())
		assert.Equal(t, DriverPostgres, parsed.Driver)
		assert.Equal(t, "photoprism", parsed.User)
		assert.Equal(t, "secret", parsed.Password)
		assert.Equal(t, "photoprism", parsed.Name)
		assert.Equal(t, "localhost", parsed.Host())
		assert.Equal(t, 5432, parsed.Port())
	})
}

func TestDSN_ParsePostgresURI(t *testing.T) {
	t.Run("Encoded", func(t *testing.T) {
		d := Parse("postgres://photo%20prism:p%40ss%3Aw%2Frd@db.local:5433/my%20db?sslmode=disable&TimeZone=UTC")
		assert.Equal(t, DriverPostgres, d.Driver)
		assert.Equal(t, "photo prism", d.User)
		assert.Equal(t, "p@ss:w/rd", d.Password)
		assert.Equal(t, "my db", d.Name)
		assert.Equal(t, "db.local:5433", d.Server)
		assert.Equal(t, "TimeZone=UTC&sslmode=disable", d.Params)
		assert.Equal(t, "db.local", d.Host())
		assert.Equal(t, 5433, d.Port())
	})
	t.Run("UnixSocket", func(t *testing.T) {
		d := Parse("postgresql://photoprism:secret@/photoprism?host=%2Fvar%2Frun%2Fpostgresql&port=5433&sslmode=disable")
		assert.Equal(t, DriverPostgres, d.Driver)
		assert.Equal(t, "/var/run/postgresql:5433", d.Server)
		assert.Equal(t, "sslmode=disable", d.Params)
		assert.Equal(t, "/var/run/postgresql", d.Host())
		assert.Equal(t, 5433, d.Port())
	})
	t.Run("RemoteHostParam", func(t *testing.T) {
		d := Parse("postgres://u:p@/db?host=db.local&sslmode=require")
		assert.Equal(t, "db.local", d.Server)
		assert.Equal(t, "sslmode=require", d.Params)
	})
	t.Run("HostAndPortParams", func(t *testing.T) {
		d := Parse("postgres://u:p@other/db?host=db.local&port=5433")
		assert.Equal(t, "db.local:5433", d.Server)
		assert.Equal(t, 5433, d.Port())
	})
	t.Run("PortParam", func(t *testing.T) {
		d := Parse("postgres://u:p@db.local/db?port=5433")
		assert.Equal(t, "db.local:5433", d.Server)
		assert.Equal(t, "db.local", d.Host())
		assert.Equal(t, 5433, d.Port())
	})
	t.Run("BuiltSocketRoundTrip", func(t *testing.T) {
		built := DSN{User: "u", Password: "p w", Server: "/run/my pg", Name: "x", Params: "port=5433"}
		d := Parse(built.PostgreSQL())
		assert.Equal(t, "p w", d.Password)
		assert.Equal(t, "/run/my pg:5433", d.Server)
	})
	t.Run("QueryCredentials", func(t *testing.T) {
		d := Parse("postgres://u:userinfo@h/db?user=q&password=param&sslmode=disable")
		assert.Equal(t, "q", d.User)
		assert.Equal(t, "param", d.Password)
		assert.Equal(t, "sslmode=disable", d.Params)
	})
	t.Run("NotURI", func(t *testing.T) {
		d := DSN{DSN: "user:secret@tcp(localhost:3306)/db"}
		assert.False(t, d.parsePostgresURI())
		d = DSN{DSN: "postgres://u:p@h:bad port/db"}
		assert.False(t, d.parsePostgresURI())
	})
}

func TestIsPostgresURI(t *testing.T) {
	assert.True(t, isPostgresURI("postgres://u:p@h/db"))
	assert.True(t, isPostgresURI("PostgreSQL://u:p@h/db"))
	assert.False(t, isPostgresURI("mysql://u:p@h/db"))
	assert.False(t, isPostgresURI("host=h options=postgres://x"))
}

//nolint:gosec // G101: masking tests use inline credential samples.
func TestMaskPostgresURI(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{name: "Encoded", in: "postgres://u:p%2Fw@h/db", out: "postgres://u:***@h/db"},
		{name: "RawAt", in: "postgres://u:p@w@h/db", out: "postgres://u:***@h/db"},
		{name: "RawSlash", in: "postgres://u:p/w@h/db", out: "postgres://u:***@h/db"},
		{name: "QueryPassword", in: "postgres://u@h/db?password=secret&sslmode=disable", out: "postgres://u@h/db?password=***&sslmode=disable"},
		{name: "Both", in: "postgresql://u:secret@h/db?password=other", out: "postgresql://u:***@h/db?password=***"},
		{name: "SslPassword", in: "postgres://u@h/db?sslpassword=key", out: "postgres://u@h/db?sslpassword=***"},
		{name: "Uppercase", in: "POSTGRES://u:SEC@RET3@h/db", out: "POSTGRES://u:***@h/db"},
		{name: "NoPassword", in: "postgres://u@h:5432/db?application_name=a", out: "postgres://u@h:5432/db?application_name=a"},
		{name: "NoPasswordAtInQuery", in: "postgres://u@h:5432/db?application_name=a@b", out: "postgres://u@h:5432/db?application_name=a@b"},
		{name: "NumericHash", in: "postgres://u:2024#TAIL9@h/db", out: "postgres://u:***@h/db"},
		{name: "NumericQuestion", in: "postgres://u:2024?TAIL10@h/db", out: "postgres://u:***@h/db"},
		{name: "HashInQueryPassword", in: "postgres://u@h/db?password=S8#TAIL8", out: "postgres://u@h/db?password=***"},
		{name: "EncodedKey", in: "postgres://u@h/db?pass%77ord=S5&x=1", out: "postgres://u@h/db?pass%77ord=***&x=1"},
		{name: "SpacedKey", in: "postgres://u@h/db?x=1& password =S6", out: "postgres://u@h/db?x=1& password =***"},
		{name: "SlashInPassword", in: "postgres://u:pa/SE%ZZ26@h/db", out: "postgres://u:***@h/db"},
		{name: "NoUserInfo", in: "postgres://h:5432/db", out: "postgres://h:5432/db"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.out, maskPostgresURI(tt.in))
		})
	}
}

//nolint:gosec // G101: masking tests use inline credential samples.
func TestMaskKeyValuePassword(t *testing.T) {
	mask := func(s string) string { return dsnKeyValuePasswordPattern.ReplaceAllStringFunc(s, maskKeyValuePassword) }

	assert.Equal(t, "host=h password=*** dbname=x", mask("host=h password=S5 dbname=x"))
	assert.Equal(t, "password='***' user=u", mask(`password='s\'e c' user=u`))
	assert.Equal(t, `password="***" user=u`, mask(`password="s e c" user=u`))
	assert.Equal(t, "user=u password=*** sslpassword=***", mask("user=u password=a://b:c@d sslpassword=k"))
	assert.Equal(t, "user=u options=x=password", mask("user=u options=x=password"))
	assert.Equal(t, "user=u password=***", mask(`user=u password=SEC\ RET16`))
	assert.Equal(t, "host=h\vpassword=***", mask("host=h\vpassword=S18"))
	assert.Equal(t, "nomatch", maskKeyValuePassword("nomatch"))
}

//nolint:gosec // G101: parser tests use inline credential samples.
func TestDSN_ParsePostgresKeyValue(t *testing.T) {
	t.Run("WithoutPassword", func(t *testing.T) {
		d := Parse("host=db.local port=5433 user=app dbname=photos sslmode=require")
		assert.Equal(t, DriverPostgres, d.Driver)
		assert.Equal(t, "app", d.User)
		assert.Equal(t, "", d.Password)
		assert.Equal(t, "photos", d.Name)
		assert.Equal(t, "db.local:5433", d.Server)
		assert.Equal(t, "sslmode=require", d.Params)
	})
	t.Run("NotKeyValue", func(t *testing.T) {
		d := DSN{DSN: "user:pass@tcp(localhost:3306)/db?charset=utf8mb4&parseTime=true"}
		assert.False(t, d.parsePostgres())
		d = DSN{DSN: "postgres://u:p@h/db?dbname=x"}
		assert.False(t, d.parsePostgres())
	})
}
