package dsn

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestValidIdent checks database name and user values.
func TestValidIdent(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		for _, s := range []string{"photoprism", "cluster_d0123456789a", "photo-prism_2", "PhotoPrism", strings.Repeat("a", 64)} {
			assert.True(t, ValidIdent(s), s)
		}
	})
	t.Run("Invalid", func(t *testing.T) {
		for _, s := range []string{"", "-x", " x", "db?x", "db/x", "db@x", "user:x", "db%41", strings.Repeat("a", 65)} {
			assert.False(t, ValidIdent(s), s)
		}
	})
}

// TestValidServer checks database server addresses.
func TestValidServer(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		for _, s := range []string{"mariadb", "mariadb:4001", ":3306", "db.example.com:3306", "db.example.com.", "10.0.0.5",
			"10.0.0.5:3306", "[::1]:3306", "::1", "photoprism_mariadb_1:3306", "mariadb-primary.photoprism.svc.cluster.local"} {
			assert.True(t, ValidServer(s), s)
		}
	})
	t.Run("Invalid", func(t *testing.T) {
		long := strings.TrimSuffix(strings.Repeat(strings.Repeat("a", 63)+".", 4), ".")
		for _, s := range []string{"", "-x", "-x:3306", "mariadb:0", "mariadb:65536", "mariadb:x", ":", "[::1", "[-x]:3306",
			"db/x", "db@x", "db?x", "tcp(db)", "db..example.com", "-db.example.com", "db-.example.com", strings.Repeat("a", 64), long,
			"/run/mysqld/mysqld.sock:3306"} {
			assert.False(t, ValidServer(s), s)
		}
	})
}

// TestValidBool checks boolean DSN parameter values.
func TestValidBool(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		for _, v := range []string{"1", "0", "true", "false", "TRUE", "False"} {
			assert.True(t, ValidBool(v), v)
		}
	})
	t.Run("Invalid", func(t *testing.T) {
		for _, v := range []string{"", "yes", "2", "tRuE", "true&x"} {
			assert.False(t, ValidBool(v), v)
		}
	})
}

// TestIsTrue checks true DSN parameter values.
func TestIsTrue(t *testing.T) {
	assert.True(t, IsTrue("1"))
	assert.True(t, IsTrue("True"))
	assert.False(t, IsTrue("false"))
	assert.False(t, IsTrue("tRuE"))
}

// TestIsFalse checks false DSN parameter values.
func TestIsFalse(t *testing.T) {
	assert.True(t, IsFalse("0"))
	assert.True(t, IsFalse("FALSE"))
	assert.False(t, IsFalse("true"))
	assert.False(t, IsFalse("fAlSe"))
}

// TestValidDuration checks duration DSN parameter values.
func TestValidDuration(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		for _, v := range []string{"15s", "1m30s", "500ms", "1.5s"} {
			assert.True(t, ValidDuration(v), v)
		}
	})
	t.Run("Invalid", func(t *testing.T) {
		for _, v := range []string{"", "0s", "-5s", "+5s", "15", "x"} {
			assert.False(t, ValidDuration(v), v)
		}
	})
}
