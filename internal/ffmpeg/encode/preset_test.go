package encode

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePreset(t *testing.T) {
	t.Run("X264Names", func(t *testing.T) {
		for _, name := range []string{PresetUltraFast, PresetSuperFast, PresetVeryFast, PresetFaster, PresetFast, PresetMedium, PresetSlow, PresetSlower, PresetVerySlow} {
			preset, ok := ParsePreset(name)
			assert.True(t, ok, name)
			assert.Equal(t, name, preset)
		}
	})
	t.Run("NvencNames", func(t *testing.T) {
		expected := map[string]string{
			"p1": PresetSuperFast,
			"p2": PresetVeryFast,
			"p3": PresetFaster,
			"p4": PresetFast,
			"p5": PresetMedium,
			"p6": PresetSlow,
			"p7": PresetSlower,
		}

		for name, want := range expected {
			preset, ok := ParsePreset(name)
			assert.True(t, ok, name)
			assert.Equal(t, want, preset, name)
		}
	})
	t.Run("X264Indexes", func(t *testing.T) {
		expected := []string{PresetUltraFast, PresetSuperFast, PresetVeryFast, PresetFaster, PresetFast, PresetMedium, PresetSlow, PresetSlower, PresetVerySlow, PresetVerySlow}

		for i, want := range expected {
			preset, ok := ParsePreset(strconv.Itoa(i))
			assert.True(t, ok, i)
			assert.Equal(t, want, preset, i)
		}
	})
	t.Run("Placebo", func(t *testing.T) {
		preset, ok := ParsePreset("placebo")
		assert.True(t, ok)
		assert.Equal(t, PresetVerySlow, preset)
	})
	t.Run("CaseAndWhitespace", func(t *testing.T) {
		preset, ok := ParsePreset(" P6 ")
		assert.True(t, ok)
		assert.Equal(t, PresetSlow, preset)
		preset, ok = ParsePreset("\tMedium\n")
		assert.True(t, ok)
		assert.Equal(t, PresetMedium, preset)
	})
	t.Run("Empty", func(t *testing.T) {
		preset, ok := ParsePreset("")
		assert.True(t, ok)
		assert.Equal(t, PresetFast, preset)
		preset, ok = ParsePreset("   ")
		assert.True(t, ok)
		assert.Equal(t, PresetFast, preset)
	})
	t.Run("Unknown", func(t *testing.T) {
		for _, name := range []string{"turbo", "p0", "p8", "10", "-1", "hq", "llhq", "fast,slow", "very fast"} {
			preset, ok := ParsePreset(name)
			assert.False(t, ok, name)
			assert.Equal(t, PresetFast, preset, name)
		}
	})
}
