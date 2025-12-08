package v4l

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

func TestV4L_TranscodeToAvcCmd_Format(t *testing.T) {
	opt := encode.NewVideoOptions("/usr/bin/ffmpeg", encode.V4LAvc, 1500, encode.DefaultQuality, encode.PresetFast, "", "0:v:0", "0:a:0?")
	cmd := TranscodeToAvcCmd("SRC.mov", "DEST.mp4", opt)
	s := cmd.String()
	assert.True(t, strings.Contains(s, "-c:v h264_v4l2m2m"))
	assert.True(t, strings.Contains(s, "-num_output_buffers 72 -num_capture_buffers 64"))
}
