package pwa

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewIcons(t *testing.T) {
	t.Run("Standard", func(t *testing.T) {
		result := NewIcons("https://demo-cdn.photoprism.app/static", "test")
		assert.NotEmpty(t, result)
		assert.Equal(t, "https://demo-cdn.photoprism.app/static/icons/test/16.png", result[0].Src)
		assert.Equal(t, "image/png", result[0].Type)
		assert.Equal(t, "16x16", result[0].Sizes)
	})
	t.Run("Custom", func(t *testing.T) {
		result := NewIcons("https://demo-cdn.photoprism.app/static", "/test.png")
		assert.NotEmpty(t, result)
		assert.Equal(t, "/test.png", result[0].Src)
		assert.Equal(t, "image/png", result[0].Type)
		assert.Equal(t, "", result[0].Sizes)
	})
}
