package pwa

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShortcutUrl(t *testing.T) {
	assert.Equal(t, "library/browse", shortcutUrl("/", "/library/", "browse"))
	assert.Equal(t, "library/settings", shortcutUrl("/instance/pro-1/", "/instance/pro-1/library", "/settings"))
	assert.Equal(t, "/library/settings", shortcutUrl("/instance/pro-1/", "/library", "/settings"))
	assert.Equal(t, "browse", shortcutUrl("/", "/", "browse"))
}

func TestStartUrl(t *testing.T) {
	t.Run("RootScope", func(t *testing.T) {
		assert.Equal(t, "./library", StartUrl("/", "/library"))
		assert.Equal(t, "./portal/admin", StartUrl("/", "/portal/admin"))
	})
	t.Run("PathScope", func(t *testing.T) {
		assert.Equal(t, "./library", StartUrl("/instance/pro-1/", "/instance/pro-1/library"))
		assert.Equal(t, "./portal/admin", StartUrl("/foo/", "/foo/portal/admin"))
	})
	t.Run("OutOfScopeFallback", func(t *testing.T) {
		assert.Equal(t, "/library", StartUrl("/instance/pro-1/", "/library"))
		assert.Equal(t, "https://example.com/library", StartUrl("/instance/pro-1/", "https://example.com/library"))
	})
}

func TestShortcuts(t *testing.T) {
	result := Shortcuts("/instance/pro-1/", "/instance/pro-1/library")

	assert.Equal(t, 4, len(result))
	assert.Equal(t, "library/browse", result[0].Url)
	assert.Equal(t, "library/albums", result[1].Url)
	assert.Equal(t, "library/places", result[2].Url)
	assert.Equal(t, "library/settings", result[3].Url)
}
