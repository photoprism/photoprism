package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBlacklists(t *testing.T) {
	t.Run("WithExtensions", func(t *testing.T) {
		lists := NewBlacklists()
		lists["foo"] = NewBlacklist("RAF, Cr3, aaf ")
		assert.True(t, lists["foo"].Contains(".raf"))
		assert.True(t, lists["foo"].Contains("cr3"))
		assert.True(t, lists["foo"].Contains("AAF"))
		assert.False(t, lists["foo"].Contains(""))
		assert.False(t, lists["foo"].Contains(".raw"))
		assert.False(t, lists["foo"].Contains("raw"))
	})
}

func TestNewBlacklist(t *testing.T) {
	t.Run("WithExtensions", func(t *testing.T) {
		list := NewBlacklist("RAF, Cr3, aaf ")
		assert.True(t, list.Contains(".raf"))
		assert.True(t, list.Contains("cr3"))
		assert.True(t, list.Contains("AAF"))
		assert.False(t, list.Contains(""))
		assert.False(t, list.Contains(".raw"))
		assert.False(t, list.Contains("raw"))
	})
}

func TestBlacklist_Ok(t *testing.T) {
	t.Run("CanonCR2", func(t *testing.T) {
		list := NewBlacklist("cr2")
		assert.False(t, list.Ok(".cr2"))
		assert.True(t, list.Contains(".cr2"))
	})
	t.Run("Raw", func(t *testing.T) {
		list := NewBlacklist("RAF, Cr3, aaf ")
		assert.False(t, list.Ok(".raf"))
		assert.False(t, list.Ok("cr3"))
		assert.False(t, list.Ok("AAF"))
		assert.True(t, list.Ok(""))
		assert.True(t, list.Ok(".raw"))
		assert.True(t, list.Ok("raw"))
	})
}
