package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUri(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		result := Uri("https://docs.photoprism.app/getting-started/config-options/#file-converters")
		assert.Equal(t, "https://docs.photoprism.app/getting-started/config-options/#file-converters", result)
	})
	t.Run("Invalid", func(t *testing.T) {
		result := Uri("https://..docs.photoprism.app/gettin\\g-started/config-options/\tfile-converters")
		assert.Equal(t, "", result)
	})
	t.Run("Emoji", func(t *testing.T) {
		result := Uri("Hello 👍")
		assert.Equal(t, "Hello%20%F0%9F%91%8D", result)
	})
}

func BenchmarkUri(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Uri("https://docs.photoprism.app/getting-started/config-options/#file-converters")
	}
}

func BenchmarkUriEmpty(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Uri("")
	}
}
