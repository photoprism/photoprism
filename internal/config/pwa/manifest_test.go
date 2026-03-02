package pwa

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewManifest(t *testing.T) {
	c := Config{
		Icon:        "logo",
		Color:       "#aaaaaa",
		Name:        "TestPrism+",
		Description: "App's Description",
		Mode:        "fullscreen",
		BaseUri:     "/",
		FrontendUri: "/library",
		StaticUri:   "/static",
	}

	t.Run("Standard", func(t *testing.T) {
		result := NewManifest(c)
		assert.NotEmpty(t, result)
		assert.Equal(t, c.Name, result.Name)
		assert.Equal(t, c.Name, result.ShortName)
		assert.Equal(t, c.Description, result.Description)
		assert.Equal(t, c.BaseUri, result.Scope)
		assert.Equal(t, StartUrl(c.BaseUri, c.FrontendUri), result.StartUrl)
		assert.Equal(t, "library/browse", result.Shortcuts[0].Url)
		assert.Len(t, result.Icons, len(IconSizes))
		assert.Len(t, result.Categories, len(Categories))
		assert.Len(t, result.Permissions, len(Permissions))
	})
	t.Run("StartUrlForPathScope", func(t *testing.T) {
		c.BaseUri = "/instance/pro-1/"
		c.FrontendUri = "/instance/pro-1/library"
		result := NewManifest(c)
		assert.NotEmpty(t, result)
		assert.Equal(t, "./library", result.StartUrl)
		assert.Equal(t, "library/browse", result.Shortcuts[0].Url)
	})
}
