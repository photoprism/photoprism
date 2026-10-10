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
