package apple

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

func TestApple_TranscodeToAvcCmd_Format(t *testing.T) {
	opt := encode.NewVideoOptions("/usr/bin/ffmpeg", encode.AppleAvc, 1500, encode.DefaultQuality, encode.PresetFast, "", "0:v:0", "0:a:0?")
	cmd := TranscodeToAvcCmd("SRC.mov", "DEST.mp4", opt)
	s := cmd.String()
	assert.True(t, strings.Contains(s, "-c:v h264_videotoolbox"))
	assert.True(t, strings.Contains(s, "-profile high -level 51 -q:v "))
	assert.True(t, strings.Contains(s, "-movflags use_metadata_tags+faststart -map_metadata 0 "))
}
