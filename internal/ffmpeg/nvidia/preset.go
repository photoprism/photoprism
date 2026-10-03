package nvidia

import (
	"strings"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

// DefaultPreset is the NVENC preset for the default "fast" encoding preset.
const DefaultPreset = "p4"

// presets maps the x264 preset names to the NVENC presets p1 (fastest) to p7 (slowest).
var presets = map[string]string{
	encode.PresetUltraFast: "p1",
	encode.PresetSuperFast: "p1",
	encode.PresetVeryFast:  "p2",
	encode.PresetFaster:    "p3",
	encode.PresetFast:      "p4",
	encode.PresetMedium:    "p5",
	encode.PresetSlow:      "p6",
	encode.PresetSlower:    "p7",
	encode.PresetVerySlow:  "p7",
	"p1":                   "p1",
	"p2":                   "p2",
	"p3":                   "p3",
	"p4":                   "p4",
	"p5":                   "p5",
	"p6":                   "p6",
	"p7":                   "p7",
}

// Preset returns the NVENC preset for the specified encoding preset name, or the default preset if it is unknown.
func Preset(name string) string {
	if preset, ok := presets[strings.ToLower(strings.TrimSpace(name))]; ok {
		return preset
	}

	return DefaultPreset
}
