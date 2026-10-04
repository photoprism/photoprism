package intel

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

func TestIntel_TranscodeToAvcCmd_WithDevice(t *testing.T) {
	opt := encode.NewVideoOptions("/usr/bin/ffmpeg", encode.IntelAvc, 1500, encode.DefaultQuality, encode.PresetFast, "/dev/dri/renderD128", "0:v:0", "0:a:0?")
	cmd := TranscodeToAvcCmd("SRC.mov", "DEST.mp4", opt)
	s := cmd.String()
	assert.True(t, strings.Contains(s, "-hwaccel qsv -hwaccel_device /dev/dri/renderD128 -hwaccel_output_format qsv"))
	assert.True(t, strings.Contains(s, "-c:v h264_qsv"))
	assert.True(t, strings.Contains(s, "-preset fast -global_quality 28"))
}

func TestIntel_TranscodeToAvcCmd_NoDevice(t *testing.T) {
	opt := encode.NewVideoOptions("/usr/bin/ffmpeg", encode.IntelAvc, 1500, encode.DefaultQuality, encode.PresetFast, "", "0:v:0", "0:a:0?")
	cmd := TranscodeToAvcCmd("SRC.mov", "DEST.mp4", opt)
	s := cmd.String()
	assert.True(t, strings.Contains(s, "-hwaccel qsv -hwaccel_output_format qsv"))
}

func TestIntel_TranscodeToAvcCmd_Preset(t *testing.T) {
	t.Run("UltraFast", func(t *testing.T) {
		opt := encode.NewVideoOptions("/usr/bin/ffmpeg", encode.IntelAvc, 1500, encode.DefaultQuality, encode.PresetUltraFast, "", "0:v:0", "0:a:0?")
		s := TranscodeToAvcCmd("SRC.mov", "DEST.mp4", opt).String()
		assert.Contains(t, s, "-preset veryfast -global_quality 28")
	})
	t.Run("UltraFastWithDevice", func(t *testing.T) {
		opt := encode.NewVideoOptions("/usr/bin/ffmpeg", encode.IntelAvc, 1500, encode.DefaultQuality, encode.PresetUltraFast, "/dev/dri/renderD128", "0:v:0", "0:a:0?")
		s := TranscodeToAvcCmd("SRC.mov", "DEST.mp4", opt).String()
		assert.Contains(t, s, "-hwaccel_device /dev/dri/renderD128")
		assert.Contains(t, s, "-preset veryfast -global_quality 28")
	})
}
