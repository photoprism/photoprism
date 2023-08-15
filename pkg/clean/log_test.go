package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLog(t *testing.T) {
	t.Run("The quick brown fox.", func(t *testing.T) {
		assert.Equal(t, "'The quick brown fox.'", Log("The quick brown fox."))
	})
	t.Run("filename.txt", func(t *testing.T) {
		assert.Equal(t, "filename.txt", Log("filename.txt"))
	})
	t.Run("empty string", func(t *testing.T) {
		assert.Equal(t, "''", Log(""))
	})
	t.Run("Replace", func(t *testing.T) {
		assert.Equal(t, "?", Log("${https://<host>:<port>/<path>}"))
	})
	t.Run("Ldap", func(t *testing.T) {
		assert.Equal(t, "?", Log("User-Agent: {jndi:ldap://<host>:<port>/<path>}"))
	})
	t.Run("SpecialChars", func(t *testing.T) {
		assert.Equal(t, "'  The ?quick? ''brown 'fox.   '", Log("  The <quick>\n\r ''brown \"fox. \t  "))
	})
}

func TestLogQuote(t *testing.T) {
	t.Run("The quick brown fox.", func(t *testing.T) {
		assert.Equal(t, "'The quick brown fox.'", LogQuote("The quick brown fox."))
	})
	t.Run("SpecialChars", func(t *testing.T) {
		assert.Equal(t, "'?The quick brown fox'", LogQuote("$The quick brown fox"))
	})
}

func TestLogLower(t *testing.T) {
	t.Run("The quick brown fox.", func(t *testing.T) {
		assert.Equal(t, "'the quick brown fox.'", LogLower("The quick brown fox."))
	})
	t.Run("filename.txt", func(t *testing.T) {
		assert.Equal(t, "filename.txt", LogLower("filename.TXT"))
	})
	t.Run("empty string", func(t *testing.T) {
		assert.Equal(t, "''", LogLower(""))
	})
	t.Run("Replace", func(t *testing.T) {
		assert.Equal(t, "?", LogLower("${https://<host>:<port>/<path>}"))
	})
	t.Run("Ldap", func(t *testing.T) {
		assert.Equal(t, "?", LogLower("User-Agent: ${jndi:ldap://<host>:<port>/<path>}"))
	})
}
