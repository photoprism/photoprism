package nvidia

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

func TestNvidia_TranscodeToAvcCmd_Format(t *testing.T) {
	opt := encode.NewVideoOptions("/usr/bin/ffmpeg", encode.NvidiaAvc, 1500, encode.DefaultQuality, encode.PresetFast, "", "0:v:0", "0:a:0?")
	cmd := TranscodeToAvcCmd("SRC.mov", "DEST.mp4", opt)
	s := cmd.String()
	assert.True(t, strings.Contains(s, "-c:v h264_nvenc"))
	assert.True(t, strings.Contains(s, "-gpu any"))
	assert.True(t, strings.Contains(s, "-rc:v constqp -cq 25"))
}
