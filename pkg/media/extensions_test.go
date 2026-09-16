package media

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMainExtensions(t *testing.T) {
	exts := MainExtensions()
	t.Run("Sorted", func(t *testing.T) {
		assert.True(t, sort.StringsAreSorted(exts))
	})
	t.Run("ContainsMainFormats", func(t *testing.T) {
		assert.Contains(t, exts, ".jpg")
		assert.Contains(t, exts, ".png")
		assert.Contains(t, exts, ".mp4")
		assert.Contains(t, exts, ".cin")
		assert.Contains(t, exts, ".ori")
		assert.Contains(t, exts, ".vob")
		assert.Contains(t, exts, ".tod")
		assert.Contains(t, exts, ".divx")
	})
	t.Run("CanonicalLowercaseOnly", func(t *testing.T) {
		assert.NotContains(t, exts, ".JPG")
	})
	t.Run("ExcludesSidecars", func(t *testing.T) {
		assert.NotContains(t, exts, ".xmp")
		assert.NotContains(t, exts, ".json")
		assert.NotContains(t, exts, ".yml")
	})
}
