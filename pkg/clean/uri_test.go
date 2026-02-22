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
	t.Run("Empty", func(t *testing.T) {
		result := Uri("")
		assert.Equal(t, "", result)
	})
}

func TestUriRedacted(t *testing.T) {
	t.Run("WithCredentials", func(t *testing.T) {
		result := UriRedacted("https://user:secret@example.com/path?q=1")
		assert.Equal(t, "https://user:xxxxx@example.com/path?q=1", result)
	})
	t.Run("WithoutCredentials", func(t *testing.T) {
		result := UriRedacted("https://docs.photoprism.app/getting-started/config-options/#file-converters")
		assert.Equal(t, "https://docs.photoprism.app/getting-started/config-options/#file-converters", result)
	})
	t.Run("Invalid", func(t *testing.T) {
		result := UriRedacted("https://..docs.photoprism.app/gettin\\g-started/config-options/\tfile-converters")
		assert.Equal(t, "", result)
	})
	t.Run("Empty", func(t *testing.T) {
		result := UriRedacted("")
		assert.Equal(t, "", result)
	})
}

func BenchmarkUri(b *testing.B) {
	for b.Loop() {
		Uri("https://docs.photoprism.app/getting-started/config-options/#file-converters")
	}
}

func BenchmarkUriRedacted(b *testing.B) {
	for b.Loop() {
		UriRedacted("https://user:secret@docs.photoprism.app/getting-started/config-options/#file-converters")
	}
}

func BenchmarkUriEmpty(b *testing.B) {
	for b.Loop() {
		Uri("")
	}
}
