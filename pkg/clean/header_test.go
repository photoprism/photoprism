package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeader(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		expected := "https://docs.photoprism.app/getting-started/config-options/#file-converters"
		result := Header(expected)
		assert.Equal(t, expected, result)
	})
	t.Run("Tabs", func(t *testing.T) {
		result := Header("https://..docs.photoprism.app/gettin\\g-started/config-options/\tfile-converters")
		assert.Equal(t, "https://..docs.photoprism.app/gettin\\g-started/config-options/file-converters", result)
	})
	t.Run("Emoji", func(t *testing.T) {
		result := Header("Hello 👍")
		assert.Equal(t, "Hello", result)
	})
}

func BenchmarkHeader(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Header("https://..docs.photoprism.app/gettin\\g-started/config-options/\tfile-converters")
	}
}

func BenchmarkHeaderEmpty(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Header("")
	}
}
