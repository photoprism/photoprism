package pwa

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestNewIcons(t *testing.T) {
	t.Run("Standard", func(t *testing.T) {
		c := Config{StaticUri: "https://demo-cdn.photoprism.app/static", StaticPath: fs.Abs("./testdata"), Icon: "test"}
		result := NewIcons(c)
		assert.NotEmpty(t, result)
		assert.Len(t, result, len(IconSizes)+len(MaskableIconSizes))
		assert.Equal(t, "https://demo-cdn.photoprism.app/static/icons/test/16.png", result[0].Src)
		assert.Equal(t, "image/png", result[0].Type)
		assert.Equal(t, "16x16", result[0].Sizes)
		assert.Equal(t, "", result[0].Purpose)
	})
	t.Run("Maskable", func(t *testing.T) {
		c := Config{StaticUri: "https://demo-cdn.photoprism.app/static", StaticPath: fs.Abs("./testdata"), Icon: "test"}
		result := NewIcons(c)
		maskable := maskableIcons(result)
		assert.Len(t, maskable, len(MaskableIconSizes))
		assert.Equal(t, "https://demo-cdn.photoprism.app/static/icons/test/maskable/192.png", maskable[0].Src)
		assert.Equal(t, "192x192", maskable[0].Sizes)
		assert.Equal(t, "image/png", maskable[0].Type)
		assert.Equal(t, "https://demo-cdn.photoprism.app/static/icons/test/maskable/512.png", maskable[1].Src)
		assert.Equal(t, "512x512", maskable[1].Sizes)
	})
	t.Run("MaskablePartial", func(t *testing.T) {
		c := Config{StaticUri: "https://demo-cdn.photoprism.app/static", StaticPath: fs.Abs("./testdata"), Icon: "partial"}
		result := NewIcons(c)
		maskable := maskableIcons(result)
		assert.Len(t, maskable, 1)
		assert.Equal(t, "https://demo-cdn.photoprism.app/static/icons/partial/maskable/192.png", maskable[0].Src)
		assert.Len(t, result, len(IconSizes)+1)
	})
	t.Run("MaskableMissing", func(t *testing.T) {
		c := Config{StaticUri: "https://demo-cdn.photoprism.app/static", StaticPath: fs.Abs("./testdata"), Icon: "nomask"}
		result := NewIcons(c)
		assert.Empty(t, maskableIcons(result))
		assert.Len(t, result, len(IconSizes))
	})
	t.Run("StaticPathUnset", func(t *testing.T) {
		c := Config{StaticUri: "https://demo-cdn.photoprism.app/static", Icon: "test"}
		result := NewIcons(c)
		assert.Empty(t, maskableIcons(result))
		assert.Len(t, result, len(IconSizes))
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

func TestMaskableIconExists(t *testing.T) {
	staticPath := fs.Abs("./testdata")
	t.Run("Exists", func(t *testing.T) {
		assert.True(t, maskableIconExists(staticPath, "test", 192))
		assert.True(t, maskableIconExists(staticPath, "test", 512))
	})
	t.Run("Missing", func(t *testing.T) {
		assert.False(t, maskableIconExists(staticPath, "test", 256))
		assert.False(t, maskableIconExists(staticPath, "partial", 512))
		assert.False(t, maskableIconExists(staticPath, "nomask", 192))
	})
	t.Run("StaticPathUnset", func(t *testing.T) {
		assert.False(t, maskableIconExists("", "test", 192))
	})
}

// maskableIcons returns the subset of icons emitted with purpose "maskable".
func maskableIcons(icons Icons) Icons {
	maskable := make(Icons, 0, len(MaskableIconSizes))
	for _, icon := range icons {
		if icon.Purpose == "maskable" {
			maskable = append(maskable, icon)
		}
	}
	return maskable
}
