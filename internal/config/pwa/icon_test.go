package pwa

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestNewIcons(t *testing.T) {
	t.Run("Standard", func(t *testing.T) {
		c := Config{StaticUri: "https://demo-cdn.photoprism.app/static", Icon: "test"}
		result := NewIcons(c)
		assert.NotEmpty(t, result)
		assert.Len(t, result, len(IconSizes)+len(MaskableIconSizes))
		assert.Equal(t, "https://demo-cdn.photoprism.app/static/icons/test/16.png", result[0].Src)
		assert.Equal(t, "image/png", result[0].Type)
		assert.Equal(t, "16x16", result[0].Sizes)
		assert.Equal(t, "", result[0].Purpose)
	})
	t.Run("Maskable", func(t *testing.T) {
		c := Config{StaticUri: "https://demo-cdn.photoprism.app/static", Icon: "test"}
		result := NewIcons(c)
		maskable := make(Icons, 0, len(MaskableIconSizes))
		for _, icon := range result {
			if icon.Purpose == "maskable" {
				maskable = append(maskable, icon)
			}
		}
		assert.Len(t, maskable, len(MaskableIconSizes))
		assert.Equal(t, "https://demo-cdn.photoprism.app/static/icons/test/maskable/192.png", maskable[0].Src)
		assert.Equal(t, "192x192", maskable[0].Sizes)
		assert.Equal(t, "image/png", maskable[0].Type)
		assert.Equal(t, "https://demo-cdn.photoprism.app/static/icons/test/maskable/512.png", maskable[1].Src)
		assert.Equal(t, "512x512", maskable[1].Sizes)
	})
	t.Run("Custom", func(t *testing.T) {
		c := Config{StaticUri: "https://demo-cdn.photoprism.app/static", Icon: "/test.png"}
		result := NewIcons(c)
		assert.NotEmpty(t, result)
		assert.Len(t, result, 1)
		assert.Equal(t, "/test.png", result[0].Src)
		assert.Equal(t, "image/png", result[0].Type)
		assert.Equal(t, "", result[0].Sizes)
		assert.Equal(t, "", result[0].Purpose)
	})
	t.Run("Theme", func(t *testing.T) {
		c := Config{StaticUri: "https://demo-cdn.photoprism.app/static", Icon: "/_theme/example.png", ThemePath: fs.Abs("./testdata"), ThemeUri: "/_theme"}
		result := NewIcons(c)
		assert.NotEmpty(t, result)
		assert.Len(t, result, 1)
		assert.Equal(t, "/_theme/example.png", result[0].Src)
		assert.Equal(t, "image/png", result[0].Type)
		assert.Equal(t, "100x67", result[0].Sizes)
		assert.Equal(t, "", result[0].Purpose)
	})
}
