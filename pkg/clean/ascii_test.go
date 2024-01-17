package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestASCII(t *testing.T) {
	t.Run("URL", func(t *testing.T) {
		result := ASCII("https://docs.photoprism.app/getting-started/config-options/#file-converters")
		assert.Equal(t, "https://docs.photoprism.app/getting-started/config-options/#file-converters", result)
	})
	t.Run("Emoji", func(t *testing.T) {
		result := ASCII("Hello 👍")
		assert.Equal(t, "Hello ", result)
	})
	t.Run("EmojiURL", func(t *testing.T) {
		result := ASCII("https://docs.photoprism.app/getting-started 👍/config-options/#file-converters")
		assert.Equal(t, "https://docs.photoprism.app/getting-started /config-options/#file-converters", result)
	})
	t.Run("Empty", func(t *testing.T) {
		result := ASCII("")
		assert.Equal(t, "", result)
	})
}

func BenchmarkASCII(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ASCII("https://docs.photoprism.app/getting-started 👍/config-options/#file-converters")
	}
}

func BenchmarkASCIIEmpty(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ASCII("")
	}
}
