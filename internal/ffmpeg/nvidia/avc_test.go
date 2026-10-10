package nvidia

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

func TestNvidia_TranscodeToAvcCmd(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		opt := encode.NewVideoOptions("/usr/bin/ffmpeg", encode.NvidiaAvc, 1500, encode.DefaultQuality, encode.PresetFast, "", "0:v:0", "0:a:0?")
		s := TranscodeToAvcCmd("SRC.mov", "DEST.mp4", opt).String()
		assert.Contains(t, s, "-c:v h264_nvenc")
		assert.Contains(t, s, "-gpu any")
		assert.Contains(t, s, " -preset p4 ")
		assert.Contains(t, s, " -rc:v vbr -cq 31 -b:v 0 -tune hq -profile:v high -level:v auto -coder:v 1 ")
		assert.NotContains(t, s, "-maxrate")
		assert.NotContains(t, s, "constqp")
	})
	t.Run("Quality", func(t *testing.T) {
		opt := encode.NewVideoOptions("/usr/bin/ffmpeg", encode.NvidiaAvc, 1500, 80, encode.PresetFast, "", "0:v:0", "0:a:0?")
		s := TranscodeToAvcCmd("SRC.mov", "DEST.mp4", opt).String()
		assert.Contains(t, s, " -cq 13 ")
	})
	t.Run("Preset", func(t *testing.T) {
		opt := encode.NewVideoOptions("/usr/bin/ffmpeg", encode.NvidiaAvc, 1500, encode.DefaultQuality, encode.PresetSlow, "", "0:v:0", "0:a:0?")
		s := TranscodeToAvcCmd("SRC.mov", "DEST.mp4", opt).String()
		assert.Contains(t, s, " -preset p6 ")
		assert.NotContains(t, s, " -preset slow ")
	})
	t.Run("MaxBitrate", func(t *testing.T) {
		opt := encode.NewVideoOptions("/usr/bin/ffmpeg", encode.NvidiaAvc, 1500, encode.DefaultQuality, encode.PresetFast, "", "0:v:0", "0:a:0?")
		opt.MaxBitrate = 25
		s := TranscodeToAvcCmd("SRC.mov", "DEST.mp4", opt).String()
		assert.Contains(t, s, " -rc:v vbr -cq 31 -b:v 0 -maxrate 25M -tune hq ")
		assert.True(t, strings.HasSuffix(s, " -f mp4 -movflags use_metadata_tags+faststart -map_metadata 0 DEST.mp4"))
	})
	t.Run("NoBitrateLimit", func(t *testing.T) {
		opt := encode.NewVideoOptions("/usr/bin/ffmpeg", encode.NvidiaAvc, 1500, encode.DefaultQuality, encode.PresetFast, "", "0:v:0", "0:a:0?")
		opt.MaxBitrate = encode.NoBitrateLimit
		s := TranscodeToAvcCmd("SRC.mov", "DEST.mp4", opt).String()
		assert.NotContains(t, s, "-maxrate")
	})
}
