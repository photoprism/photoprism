package encode

import (
	"strings"
)

// FFmpeg encoding preset names from fastest to slowest,
// see https://trac.ffmpeg.org/wiki/Encode/H.264#Preset.
const (
	PresetUltraFast = "ultrafast"
	PresetSuperFast = "superfast"
	PresetVeryFast  = "veryfast"
	PresetFaster    = "faster"
	PresetFast      = "fast"
	PresetMedium    = "medium"
	PresetSlow      = "slow"
	PresetSlower    = "slower"
	PresetVerySlow  = "veryslow"
)

// presets maps the accepted preset names to the x264 preset names that every encoder can map.
// NVENC names select the x264 preset of the same speed, the x264 index strings "0" to "9" the
// preset they number, and "placebo" the slowest common one.
var presets = map[string]string{
	PresetUltraFast: PresetUltraFast,
	PresetSuperFast: PresetSuperFast,
	PresetVeryFast:  PresetVeryFast,
	PresetFaster:    PresetFaster,
	PresetFast:      PresetFast,
	PresetMedium:    PresetMedium,
	PresetSlow:      PresetSlow,
	PresetSlower:    PresetSlower,
	PresetVerySlow:  PresetVerySlow,
	"placebo":       PresetVerySlow,
	"0":             PresetUltraFast,
	"1":             PresetSuperFast,
	"2":             PresetVeryFast,
	"3":             PresetFaster,
	"4":             PresetFast,
	"5":             PresetMedium,
	"6":             PresetSlow,
	"7":             PresetSlower,
	"8":             PresetVerySlow,
	"9":             PresetVerySlow,
	"p1":            PresetSuperFast,
	"p2":            PresetVeryFast,
	"p3":            PresetFaster,
	"p4":            PresetFast,
	"p5":            PresetMedium,
	"p6":            PresetSlow,
	"p7":            PresetSlower,
}

// ParsePreset returns the x264 preset name for the specified encoding preset, ignoring case and
// surrounding whitespace. It returns PresetFast for an empty name, and false if the name is unknown.
func ParsePreset(name string) (preset string, ok bool) {
	name = strings.ToLower(strings.TrimSpace(name))

	if name == "" {
		return PresetFast, true
	} else if preset, ok = presets[name]; ok {
		return preset, true
	}

	return PresetFast, false
}
