package intel

import (
	"strings"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

// DefaultPreset is the Quick Sync preset for unknown encoding preset names.
const DefaultPreset = encode.PresetFast

// presets maps the x264 preset names to the Quick Sync presets, which range from "veryfast" to "veryslow".
var presets = map[string]string{
	encode.PresetUltraFast: encode.PresetVeryFast,
	encode.PresetSuperFast: encode.PresetVeryFast,
	encode.PresetVeryFast:  encode.PresetVeryFast,
	encode.PresetFaster:    encode.PresetFaster,
	encode.PresetFast:      encode.PresetFast,
	encode.PresetMedium:    encode.PresetMedium,
	encode.PresetSlow:      encode.PresetSlow,
	encode.PresetSlower:    encode.PresetSlower,
	encode.PresetVerySlow:  encode.PresetVerySlow,
}

// Preset returns the Quick Sync preset for the specified encoding preset name, or the default preset if it is unknown.
func Preset(name string) string {
	if preset, ok := presets[strings.ToLower(strings.TrimSpace(name))]; ok {
		return preset
	}

	return DefaultPreset
}
