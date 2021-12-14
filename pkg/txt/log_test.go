package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogParam(t *testing.T) {
	t.Run("The quick brown fox.", func(t *testing.T) {
		assert.Equal(t, "'The quick brown fox.'", LogParam("The quick brown fox."))
	})
	t.Run("filename.txt", func(t *testing.T) {
		assert.Equal(t, "filename.txt", LogParam("filename.txt"))
	})
	t.Run("empty string", func(t *testing.T) {
		assert.Equal(t, "''", LogParam(""))
	})
	t.Run("Log4j", func(t *testing.T) {
		assert.Equal(t, "??https://?host?:?port?/?path??", LogParam("${https://<host>:<port>/<path>}"))
	})
	t.Run("Log4j", func(t *testing.T) {
		assert.Equal(t, "?", LogParam("User-Agent: ${jndi:ldap://<host>:<port>/<path>}"))
	})
}

func TestLogParamLower(t *testing.T) {
	t.Run("The quick brown fox.", func(t *testing.T) {
		assert.Equal(t, "'the quick brown fox.'", LogParamLower("The quick brown fox."))
	})
	t.Run("filename.txt", func(t *testing.T) {
		assert.Equal(t, "filename.txt", LogParamLower("filename.TXT"))
	})
	t.Run("empty string", func(t *testing.T) {
		assert.Equal(t, "''", LogParamLower(""))
	})
	t.Run("Log4j", func(t *testing.T) {
		assert.Equal(t, "??https://?host?:?port?/?path??", LogParamLower("${https://<host>:<port>/<path>}"))
	})
	t.Run("Log4j", func(t *testing.T) {
		assert.Equal(t, "?", LogParamLower("User-Agent: ${jndi:ldap://<host>:<port>/<path>}"))
	})
}
