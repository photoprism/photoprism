package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewExtLists(t *testing.T) {
	t.Run("WithExtensions", func(t *testing.T) {
		lists := NewExtLists()
		lists["foo"] = NewExtList("RAF, Cr3, aaf ")
		assert.True(t, lists["foo"].Contains(".raf"))
		assert.True(t, lists["foo"].Contains("cr3"))
		assert.True(t, lists["foo"].Contains("AAF"))
		assert.False(t, lists["foo"].Contains(""))
		assert.False(t, lists["foo"].Contains(".raw"))
		assert.False(t, lists["foo"].Contains("raw"))
	})
}

func TestNewExtList(t *testing.T) {
	t.Run("WithExtensions", func(t *testing.T) {
		list := NewExtList("RAF, Cr3, aaf ")
		assert.True(t, list.Contains(".raf"))
		assert.True(t, list.Contains("cr3"))
		assert.True(t, list.Contains("AAF"))
		assert.False(t, list.Contains(""))
		assert.False(t, list.Contains(".raw"))
		assert.False(t, list.Contains("raw"))
	})
}

func TestExtList_Ok(t *testing.T) {
	t.Run("CanonCR2", func(t *testing.T) {
		list := NewExtList("cr2")
		assert.False(t, list.Allow(".cr2"))
		assert.True(t, list.Contains(".cr2"))
	})
	t.Run("Raw", func(t *testing.T) {
		list := NewExtList("RAF, Cr3, aaf ")
		assert.False(t, list.Allow(".raf"))
		assert.False(t, list.Allow("cr3"))
		assert.False(t, list.Allow("AAF"))
		assert.True(t, list.Allow(""))
		assert.True(t, list.Allow(".raw"))
		assert.True(t, list.Allow("raw"))
	})
}

func TestExtList_Contains(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		list := NewExtList("dng")
		assert.True(t, list.Contains("dng"))
		assert.False(t, list.Contains("cr2"))
	})
	t.Run("Empty", func(t *testing.T) {
		list := NewExtList("")
		assert.False(t, list.Contains(""))
	})
}

func TestExtList_Excludes(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		list := NewExtList("dng")
		assert.False(t, list.Excludes("dng"))
		assert.True(t, list.Excludes("cr2"))
	})
	t.Run("Empty", func(t *testing.T) {
		list := NewExtList("")
		assert.False(t, list.Excludes(""))
		assert.False(t, list.Excludes("dng"))
		assert.False(t, list.Excludes("cr2"))
	})
}

func TestExtList_Set(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		list := NewExtList("dng")
		assert.True(t, list.Contains("dng"))
		assert.False(t, list.Contains("cr2"))
		list.Set("cr2")
		assert.True(t, list.Contains("dng"))
		assert.True(t, list.Contains("cr2"))
	})
	t.Run("Empty", func(t *testing.T) {
		list := NewExtList("dng")
		assert.True(t, list.Contains("dng"))
		assert.False(t, list.Contains("cr2"))
		list.Set("")
		assert.True(t, list.Contains("dng"))
		assert.False(t, list.Contains("cr2"))
	})
}

func TestExtList_Add(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		list := NewExtList("dng")
		assert.True(t, list.Contains("dng"))
		assert.False(t, list.Contains("cr2"))
		list.Add("cr2")
		assert.True(t, list.Contains("dng"))
		assert.True(t, list.Contains("cr2"))
	})
	t.Run("Empty", func(t *testing.T) {
		list := NewExtList("dng")
		assert.True(t, list.Contains("dng"))
		assert.False(t, list.Contains("cr2"))
		list.Add("")
		assert.True(t, list.Contains("dng"))
		assert.False(t, list.Contains("cr2"))
	})
}

func TestExtList_String(t *testing.T) {
	t.Run("One", func(t *testing.T) {
		list := NewExtList("jpg")
		assert.Equal(t, "jpg", list.String())
	})
	t.Run("Two", func(t *testing.T) {
		list := NewExtList("dng, CR2")
		assert.Equal(t, "cr2, dng", list.String())
	})
	t.Run("Empty", func(t *testing.T) {
		list := NewExtList("")
		assert.Equal(t, "", list.String())
	})
}

func TestExtList_Accept(t *testing.T) {
	t.Run("One", func(t *testing.T) {
		list := NewExtList("jpg")
		assert.Equal(t, ".jpg", list.Accept())
	})
	t.Run("Two", func(t *testing.T) {
		list := NewExtList("mp4, avi")
		assert.Equal(t, ".avi,.mp4", list.Accept())
	})
	t.Run("Jpeg", func(t *testing.T) {
		list := NewExtList("jpg,jpeg,jxl")
		assert.Equal(t, ".jpeg,.jpg,.jxl", list.Accept())
	})
	t.Run("Empty", func(t *testing.T) {
		list := NewExtList("")
		assert.Equal(t, "", list.Accept())
	})
}
