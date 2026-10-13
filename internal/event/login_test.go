package event

import (
	"strings"
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
			"Ab\x00c\x01d\x07e\x07f\r", "a\x7fb\u202Ec", "line\nnext\x00")
		assert.Equal(t, "192.0.2.1[U+0000]", data["ip"])
		assert.Equal(t, "web[U+001B]dav", data["realm"])
		assert.Equal(t, "Ab[U+0000]c[U+0001]d[U+0007]e[U+0007]f", data["name"])
		assert.Equal(t, "a[U+007F]b[U+202E]c", data["browser"])
		assert.Equal(t, "line next[U+0000]", data["message"])
	})
	t.Run("LongName", func(t *testing.T) {
		data := LoginData(logrus.ErrorLevel, "192.0.2.1", "api", strings.Repeat("\u200B", 30), "", "")
		assert.Equal(t, strings.Repeat("[U+200B]", 8), data["name"])
		data = LoginData(logrus.ErrorLevel, "192.0.2.1", "api", "abc"+strings.Repeat("\u200B", 30), "", "")
		assert.Equal(t, "abc"+strings.Repeat("[U+200B]", 7), data["name"])
	})
}
