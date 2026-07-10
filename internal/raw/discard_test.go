package raw

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiscardRenderOnWarning(t *testing.T) {
	// Pin the gated set so the predicate is independent of any process-level env override.
	orig := discardRenderOnWarning
	discardRenderOnWarning = parseDiscardExt(".cr3")
	t.Cleanup(func() { discardRenderOnWarning = orig })

	t.Run("Gated", func(t *testing.T) {
		assert.True(t, DiscardRenderOnWarning(".cr3"))
	})
	t.Run("NotGated", func(t *testing.T) {
		assert.False(t, DiscardRenderOnWarning(".raw"))
		assert.False(t, DiscardRenderOnWarning(".kdc"))
		assert.False(t, DiscardRenderOnWarning(".cr2"))
		assert.False(t, DiscardRenderOnWarning(".dng"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, DiscardRenderOnWarning(""))
	})
}

func TestDefaultDiscardExt(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		set := defaultDiscardExt()
		assert.True(t, set[".cr3"])
		assert.Len(t, set, 1)
	})
	t.Run("EnvOverrideReplaces", func(t *testing.T) {
		// The override replaces the default set rather than extending it.
		t.Setenv(discardOnWarningEnv, "cr2, kdc")
		set := defaultDiscardExt()
		assert.True(t, set[".cr2"])
		assert.True(t, set[".kdc"])
		assert.False(t, set[".cr3"])
	})
	t.Run("EnvBlankFallsBack", func(t *testing.T) {
		t.Setenv(discardOnWarningEnv, "   ")
		set := defaultDiscardExt()
		assert.True(t, set[".cr3"])
		assert.Len(t, set, 1)
	})
	t.Run("EnvDisablesGate", func(t *testing.T) {
		// A comma-only value yields no valid extension and disables the gate.
		t.Setenv(discardOnWarningEnv, ",")
		assert.Empty(t, defaultDiscardExt())
	})
}

func TestParseDiscardExt(t *testing.T) {
	t.Run("AddsLeadingDot", func(t *testing.T) {
		set := parseDiscardExt("cr3")
		assert.True(t, set[".cr3"])
	})
	t.Run("KeepsLeadingDot", func(t *testing.T) {
		set := parseDiscardExt(".cr3")
		assert.True(t, set[".cr3"])
	})
	t.Run("TrimsAndLowercases", func(t *testing.T) {
		set := parseDiscardExt("  .CR3 , KDC ")
		assert.True(t, set[".cr3"])
		assert.True(t, set[".kdc"])
		assert.Len(t, set, 2)
	})
	t.Run("SkipsBlankEntries", func(t *testing.T) {
		set := parseDiscardExt("cr3, , ,cr2")
		assert.True(t, set[".cr3"])
		assert.True(t, set[".cr2"])
		assert.Len(t, set, 2)
	})
	t.Run("EmptyInput", func(t *testing.T) {
		assert.Empty(t, parseDiscardExt(""))
	})
}

func TestDiscardExtMissingPreview(t *testing.T) {
	orig := discardRenderOnWarning
	t.Cleanup(func() { discardRenderOnWarning = orig })

	t.Run("ShippedDefaultIsSafe", func(t *testing.T) {
		// The default gated set must never include a preview-unsafe format.
		for ext := range defaultDiscardExt() {
			assert.Truef(t, PreviewExtAllowed(ext), "default gated format %s must allow preview extraction", ext)
		}
	})
	t.Run("DefaultClean", func(t *testing.T) {
		discardRenderOnWarning = parseDiscardExt(".cr3")
		assert.Empty(t, discardExtMissingPreview())
	})
	t.Run("DetectsPreviewUnsafe", func(t *testing.T) {
		// A preview-unsafe override (.mos) would discard the render with no fallback.
		discardRenderOnWarning = parseDiscardExt(".cr3, .mos")
		assert.Equal(t, []string{".mos"}, discardExtMissingPreview())
	})
}
