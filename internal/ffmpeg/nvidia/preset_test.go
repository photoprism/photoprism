package nvidia

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

func TestPreset(t *testing.T) {
	t.Run("X264Names", func(t *testing.T) {
		assert.Equal(t, "p1", Preset(encode.PresetUltraFast))
		assert.Equal(t, "p1", Preset(encode.PresetSuperFast))
		assert.Equal(t, "p2", Preset(encode.PresetVeryFast))
		assert.Equal(t, "p3", Preset(encode.PresetFaster))
		assert.Equal(t, "p4", Preset(encode.PresetFast))
		assert.Equal(t, "p5", Preset(encode.PresetMedium))
		assert.Equal(t, "p6", Preset(encode.PresetSlow))
		assert.Equal(t, "p7", Preset(encode.PresetSlower))
		assert.Equal(t, "p7", Preset(encode.PresetVerySlow))
	})
	t.Run("Passthrough", func(t *testing.T) {
		for _, p := range []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7"} {
			assert.Equal(t, p, Preset(p))
		}
		assert.Equal(t, "p6", Preset(" P6 "))
		assert.Equal(t, "p5", Preset("Medium"))
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.Equal(t, "p4", Preset(""))
		assert.Equal(t, DefaultPreset, Preset(""))
		assert.Equal(t, DefaultPreset, Preset("p0"))
		assert.Equal(t, DefaultPreset, Preset("p8"))
		assert.Equal(t, DefaultPreset, Preset("hq"))
		assert.Equal(t, DefaultPreset, Preset("llhq"))
		assert.Equal(t, DefaultPreset, Preset("placebo"))
		assert.Equal(t, DefaultPreset, Preset("fast -x"))
	})
	t.Run("DefaultPreset", func(t *testing.T) {
		assert.Equal(t, "p4", DefaultPreset)
		assert.Equal(t, presets[encode.PresetFast], DefaultPreset)
	})
}
