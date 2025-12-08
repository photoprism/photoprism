package vaapi

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

func TestVaapi_TranscodeToAvcCmd_WithDevice(t *testing.T) {
	opt := encode.NewVideoOptions("/usr/bin/ffmpeg", encode.VaapiAvc, 1500, encode.DefaultQuality, encode.PresetFast, "/dev/dri/renderD128", "0:v:0", "0:a:0?")
	cmd := TranscodeToAvcCmd("SRC.mov", "DEST.mp4", opt)
	s := cmd.String()
	assert.True(t, strings.Contains(s, "-hwaccel vaapi -hwaccel_device /dev/dri/renderD128"))
	assert.True(t, strings.Contains(s, "-c:v h264_vaapi"))
	assert.True(t, strings.Contains(s, "-qp 25"))
}

func TestVaapi_TranscodeToAvcCmd_NoDevice(t *testing.T) {
	opt := encode.NewVideoOptions("/usr/bin/ffmpeg", encode.VaapiAvc, 1500, encode.DefaultQuality, encode.PresetFast, "", "0:v:0", "0:a:0?")
	cmd := TranscodeToAvcCmd("SRC.mov", "DEST.mp4", opt)
	s := cmd.String()
	assert.True(t, strings.Contains(s, "-hwaccel vaapi"))
}
