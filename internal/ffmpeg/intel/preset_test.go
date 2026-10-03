package intel

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

func TestPreset(t *testing.T) {
	t.Run("X264Names", func(t *testing.T) {
		assert.Equal(t, "veryfast", Preset(encode.PresetUltraFast))
		assert.Equal(t, "veryfast", Preset(encode.PresetSuperFast))
		assert.Equal(t, "veryfast", Preset(encode.PresetVeryFast))
		assert.Equal(t, "faster", Preset(encode.PresetFaster))
		assert.Equal(t, "fast", Preset(encode.PresetFast))
		assert.Equal(t, "medium", Preset(encode.PresetMedium))
		assert.Equal(t, "slow", Preset(encode.PresetSlow))
		assert.Equal(t, "slower", Preset(encode.PresetSlower))
		assert.Equal(t, "veryslow", Preset(encode.PresetVerySlow))
	})
	t.Run("CaseAndWhitespace", func(t *testing.T) {
		assert.Equal(t, "medium", Preset(" Medium "))
		assert.Equal(t, "veryfast", Preset("ULTRAFAST"))
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.Equal(t, DefaultPreset, Preset(""))
		assert.Equal(t, DefaultPreset, Preset("p6"))
		assert.Equal(t, DefaultPreset, Preset("placebo"))
		assert.Equal(t, DefaultPreset, Preset("turbo"))
	})
	t.Run("DefaultPreset", func(t *testing.T) {
		assert.Equal(t, "fast", DefaultPreset)
	})
}
