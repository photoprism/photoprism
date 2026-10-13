package event

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLoginData(t *testing.T) {
	t.Run("Unchanged", func(t *testing.T) {
		ua := "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0 Safari/537.36"
		data := LoginData(logrus.ErrorLevel, "192.0.2.1", "webdav", "alice", ua, "invalid credentials")
		assert.Equal(t, "error", data["level"])
		assert.Equal(t, "192.0.2.1", data["ip"])
		assert.Equal(t, "webdav", data["realm"])
		assert.Equal(t, "alice", data["name"])
		assert.Equal(t, ua, data["browser"])
		assert.Equal(t, "invalid credentials", data["message"])
	})
	t.Run("NonLatin", func(t *testing.T) {
		data := LoginData(logrus.InfoLevel, "2001:db8::1", "api", "مستخدم", "日本語", "")
		assert.Equal(t, "مستخدم", data["name"])
		assert.Equal(t, "日本語", data["browser"])
	})
	t.Run("ControlCharacters", func(t *testing.T) {
		data := LoginData(logrus.ErrorLevel, "192.0.2.1\x00", "web\x1bdav",
			"Ab\x00c\x01d\x1b[31me\x07f\r", "a\x7fb\u202Ec", "line\nnext\x00")
		assert.Equal(t, "192.0.2.1", data["ip"])
		assert.Equal(t, "webdav", data["realm"])
		assert.Equal(t, "Abcd[31mef", data["name"])
		assert.Equal(t, "ab?c", data["browser"])
		assert.Equal(t, "line next", data["message"])
	})
}
