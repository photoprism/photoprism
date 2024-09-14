package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/txt/clip"
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
	t.Run("LoremIpsum", func(t *testing.T) {
		assert.Equal(t, "'It is a long established fact that a reader will be distracted by the readable "+
			"content of a pagewhen looking at its layout. The point of using Lorem Ipsum is that it has a "+
			"more-or-less normal distribution of letters,as opposed to using 'Content here, content here', making it "+
			"look like readable English.Many desktop publishing packages and web page editors now use Lorem Ipsum as "+
			"their default model text, and a search for'lorem ipsum' will uncover many web sites still in their "+
			"infancy. Various versions…'", Log(clip.LoremIpsum))
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
