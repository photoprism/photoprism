package video

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFormats(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		list := NewFormats("")
		assert.Empty(t, list)
	})
	t.Run("Single", func(t *testing.T) {
		list := NewFormats("magicyuv")
		assert.Len(t, list, 1)
		assert.True(t, list.Contains("magicyuv"))
	})
	t.Run("CommaSeparated", func(t *testing.T) {
		list := NewFormats("magicyuv, hap, avi")
		assert.Len(t, list, 3)
		assert.True(t, list.Contains("magicyuv"))
		assert.True(t, list.Contains("hap"))
		assert.True(t, list.Contains("avi"))
	})
	t.Run("AliasesCollapse", func(t *testing.T) {
		// Different names for the same codec map to a single canonical entry.
		list := NewFormats("magicyuv, m8rg ,M8RA")
		assert.Len(t, list, 1)
		assert.True(t, list.Contains("magy"))
		assert.True(t, list.Contains("magicyuv"))
		assert.True(t, list.Contains("m8rg"))
		assert.True(t, list.Contains("m8ra"))
	})
	t.Run("MixedContainerAndCodec", func(t *testing.T) {
		list := NewFormats("avi, magicyuv")
		assert.Len(t, list, 2)
		assert.True(t, list.Contains("avi"))
		assert.True(t, list.Contains("magicyuv"))
	})
}

func TestFormats_Contains(t *testing.T) {
	list := NewFormats("magicyuv, m8rg")

	t.Run("Match", func(t *testing.T) {
		assert.True(t, list.Contains("magicyuv"))
	})
	t.Run("CaseInsensitive", func(t *testing.T) {
		assert.True(t, list.Contains("MagicYUV"))
		assert.True(t, list.Contains("M8RG"))
	})
	t.Run("Trimmed", func(t *testing.T) {
		assert.True(t, list.Contains("  magicyuv  "))
		assert.True(t, list.Contains(`"magicyuv"`))
		assert.True(t, list.Contains(".magicyuv"))
	})
	t.Run("NoMatch", func(t *testing.T) {
		assert.False(t, list.Contains("avc1"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, list.Contains(""))
	})
	t.Run("EmptyList", func(t *testing.T) {
		empty := NewFormats("")
		assert.False(t, empty.Contains("magicyuv"))
	})
	t.Run("NoArgs", func(t *testing.T) {
		assert.False(t, list.Contains())
	})
	t.Run("MultipleArgsFirstMatches", func(t *testing.T) {
		assert.True(t, list.Contains("magicyuv", "avi"))
	})
	t.Run("MultipleArgsSecondMatches", func(t *testing.T) {
		assert.True(t, list.Contains("avc1", "magicyuv"))
	})
	t.Run("MultipleArgsWithEmpty", func(t *testing.T) {
		assert.True(t, list.Contains("", "magicyuv"))
		assert.False(t, list.Contains("", "avc1"))
	})
	t.Run("MultipleArgsNoMatch", func(t *testing.T) {
		assert.False(t, list.Contains("avc1", "hevc", "mp4"))
	})
	t.Run("CodecAliasOriginalAndMapped", func(t *testing.T) {
		// A list built from the canonical name matches the original FourCC and
		// the human-readable alias reported for the same codec.
		excl := NewFormats("magy")
		assert.True(t, excl.Contains("magy"))     // Canonical name.
		assert.True(t, excl.Contains("magicyuv")) // Human-readable alias.
		assert.True(t, excl.Contains("m8ra"))     // FourCC reported by ExifTool.
		assert.True(t, excl.Contains("M8RG"))     // FourCC, case-insensitive.
		assert.False(t, excl.Contains("avc1"))
	})
	t.Run("CodecAliasReverse", func(t *testing.T) {
		// A list built from an alias still matches the canonical name and other
		// aliases for the same codec.
		excl := NewFormats("magicyuv")
		assert.True(t, excl.Contains("magy"))
		assert.True(t, excl.Contains("m8ra"))
	})
	t.Run("H264Alias", func(t *testing.T) {
		excl := NewFormats("h264")
		assert.True(t, excl.Contains("avc1"))
		assert.True(t, excl.Contains("avc"))
	})
	t.Run("VFWWrapper", func(t *testing.T) {
		// Excluding the VFW wrapper matches the Matroska codec ID reported for it.
		excl := NewFormats("vfw")
		assert.True(t, excl.Contains("V_MS/VFW/FOURCC"))
		assert.True(t, excl.Contains("v_ms"))
		assert.False(t, excl.Contains("avc1"))
	})
}

func TestFormats_Allow(t *testing.T) {
	list := NewFormats("magicyuv")

	assert.False(t, list.Allow("magicyuv"))
	assert.True(t, list.Allow("avc1"))
	assert.True(t, list.Allow(""))
}

func TestFormats_Match(t *testing.T) {
	list := NewFormats("magy, mov")

	t.Run("ReturnsCanonicalEntry", func(t *testing.T) {
		// The matched entry is reported, not the first argument.
		assert.Equal(t, "mov", list.Match("avc1", "mov"))
		assert.Equal(t, "magy", list.Match("m8ra", "avi"))
	})
	t.Run("SkipsEmptyArgs", func(t *testing.T) {
		assert.Equal(t, "magy", list.Match("", "magicyuv"))
	})
	t.Run("NoMatch", func(t *testing.T) {
		assert.Equal(t, "", list.Match("avc1", "mp4"))
		assert.Equal(t, "", list.Match())
		assert.Equal(t, "", list.Match(""))
	})
	t.Run("EmptyList", func(t *testing.T) {
		assert.Equal(t, "", NewFormats("").Match("magy"))
	})
}

func TestFormats_Set(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		list := make(Formats)
		list.Set("")
		assert.Empty(t, list)
	})
	t.Run("AddsToExisting", func(t *testing.T) {
		list := NewFormats("magicyuv")
		list.Set("m8rg, AVI")
		assert.Len(t, list, 2) // m8rg collapses into the existing magy entry.
		assert.True(t, list.Contains("magicyuv"))
		assert.True(t, list.Contains("m8rg"))
		assert.True(t, list.Contains("avi"))
	})
	t.Run("Duplicate", func(t *testing.T) {
		list := NewFormats("magicyuv")
		list.Set("MAGICYUV, magicyuv")
		assert.Len(t, list, 1)
	})
}

func TestFormats_Add(t *testing.T) {
	t.Run("Single", func(t *testing.T) {
		list := make(Formats)
		list.Add("MagicYUV")
		assert.Len(t, list, 1)
		assert.True(t, list.Contains("magicyuv"))
	})
	t.Run("Empty", func(t *testing.T) {
		list := make(Formats)
		list.Add("")
		list.Add("   ")
		list.Add(".")
		assert.Empty(t, list)
	})
	t.Run("Idempotent", func(t *testing.T) {
		list := make(Formats)
		list.Add("avi")
		list.Add("avi")
		assert.Len(t, list, 1)
	})
}

func TestFormats_String(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		list := NewFormats("")
		assert.Equal(t, "", list.String())
	})
	t.Run("Sorted", func(t *testing.T) {
		list := NewFormats("magy, hap, avi")
		assert.Equal(t, "avi, hap, magy", list.String())
	})
	t.Run("CanonicalForm", func(t *testing.T) {
		// Codec aliases are stored and reported in their canonical form.
		list := NewFormats("magicyuv, m8ra")
		assert.Equal(t, "magy", list.String())
	})
}
