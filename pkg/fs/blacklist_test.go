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
		assert.False(t, list.Allow(".cr2"))
		assert.True(t, list.Contains(".cr2"))
	})
	t.Run("Raw", func(t *testing.T) {
		list := NewBlacklist("RAF, Cr3, aaf ")
		assert.False(t, list.Allow(".raf"))
		assert.False(t, list.Allow("cr3"))
		assert.False(t, list.Allow("AAF"))
		assert.True(t, list.Allow(""))
		assert.True(t, list.Allow(".raw"))
		assert.True(t, list.Allow("raw"))
	})
}

func TestBlacklist_Contains(t *testing.T) {
	t.Run("DNG", func(t *testing.T) {
		list := NewBlacklist("dng")
		assert.True(t, list.Contains("dng"))
		assert.False(t, list.Contains("cr2"))
	})
	t.Run("Empty", func(t *testing.T) {
		list := NewBlacklist("")
		assert.False(t, list.Contains(""))
	})
}

func TestBlacklist_Set(t *testing.T) {
	t.Run("DNG, CR2", func(t *testing.T) {
		list := NewBlacklist("dng")
		assert.True(t, list.Contains("dng"))
		assert.False(t, list.Contains("cr2"))
		list.Set("cr2")
		assert.True(t, list.Contains("dng"))
		assert.True(t, list.Contains("cr2"))
	})
	t.Run("DNG", func(t *testing.T) {
		list := NewBlacklist("dng")
		assert.True(t, list.Contains("dng"))
		assert.False(t, list.Contains("cr2"))
		list.Set("")
		assert.True(t, list.Contains("dng"))
		assert.False(t, list.Contains("cr2"))
	})
}

func TestBlacklist_Add(t *testing.T) {
	t.Run("DNG, CR2", func(t *testing.T) {
		list := NewBlacklist("dng")
		assert.True(t, list.Contains("dng"))
		assert.False(t, list.Contains("cr2"))
		list.Add("cr2")
		assert.True(t, list.Contains("dng"))
		assert.True(t, list.Contains("cr2"))
	})
	t.Run("DNG", func(t *testing.T) {
		list := NewBlacklist("dng")
		assert.True(t, list.Contains("dng"))
		assert.False(t, list.Contains("cr2"))
		list.Add("")
		assert.True(t, list.Contains("dng"))
		assert.False(t, list.Contains("cr2"))
	})
}
