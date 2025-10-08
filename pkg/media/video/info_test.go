package video

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/service/http/header"
)

func TestInfo(t *testing.T) {
	t.Run("VideoSize", func(t *testing.T) {
		info := NewInfo()
		info.FileSize = 1005000
		info.VideoOffset = 5000
		assert.Equal(t, int64(1000000), info.VideoSize())
	})
	t.Run("VideoBitrate", func(t *testing.T) {
		info := NewInfo()
		info.FileSize = 1005000
		info.VideoOffset = 5000
		info.Duration = time.Second
		assert.Equal(t, float64(8), info.VideoBitrate())
	})
	t.Run("VideoContentType", func(t *testing.T) {
		info := NewInfo()
		info.VideoMimeType = header.ContentTypeMp4
		info.VideoCodec = CodecAvc1
		assert.Equal(t, header.ContentTypeMp4AvcMain, info.VideoContentType())
	})
	t.Run("VideoFileExt", func(t *testing.T) {
		info := NewInfo()
		info.VideoMimeType = header.ContentTypeMp4
		info.VideoCodec = CodecAvc1
		assert.Equal(t, fs.ExtMp4, info.VideoFileExt())
	})
	t.Run("VideoFileType", func(t *testing.T) {
		info := NewInfo()
		info.VideoMimeType = header.ContentTypeMp4
		info.VideoCodec = CodecAvc1
		assert.Equal(t, fs.VideoMp4, info.VideoFileType())
	})
}

func TestInfo_VideoSize(t *testing.T) {
	// Negative values yield 0
	assert.Equal(t, int64(0), Info{FileSize: -1, VideoOffset: 0}.VideoSize())
	assert.Equal(t, int64(0), Info{FileSize: 10, VideoOffset: -1}.VideoSize())
	// Normal size
	assert.Equal(t, int64(90), Info{FileSize: 100, VideoOffset: 10}.VideoSize())
}

func TestInfo_VideoBitrate(t *testing.T) {
	// Unknown size or duration yields 0
	assert.Equal(t, 0.0, Info{FileSize: -1, VideoOffset: 0, Duration: time.Second}.VideoBitrate())
	assert.Equal(t, 0.0, Info{FileSize: 100, VideoOffset: 50, Duration: 0}.VideoBitrate())
	// Bitrate: (size*8)/(duration) in Mbps
	inf := Info{FileSize: 1000, VideoOffset: 500, Duration: time.Second}
	// size = 500 bytes; bitrate = (500*8)/1 / 1e6 = 0.004 Mbps
	assert.InDelta(t, 0.004, inf.VideoBitrate(), 1e-6)
}

func TestInfo_VideoFileExtAndType(t *testing.T) {
	// MOV maps to .mov and VideoMov
	mov := Info{VideoMimeType: header.ContentTypeMov}
	if got := mov.VideoFileExt(); got != fs.ExtMov {
		t.Fatalf("mov ext: got=%s want=%s", got, fs.ExtMov)
	}
	if got := mov.VideoFileType(); got != fs.VideoMov {
		t.Fatalf("mov type: got=%v want=%v", got, fs.VideoMov)
	}

	// MP4 maps to .mp4 and VideoMp4
	mp4 := Info{VideoMimeType: header.ContentTypeMp4}
	if got := mp4.VideoFileExt(); got != fs.ExtMp4 {
		t.Fatalf("mp4 ext: got=%s want=%s", got, fs.ExtMp4)
	}
	if got := mp4.VideoFileType(); got != fs.VideoMp4 {
		t.Fatalf("mp4 type: got=%v want=%v", got, fs.VideoMp4)
	}

	// Unknown defaults to MP4
	unk := Info{VideoMimeType: ""}
	if got := unk.VideoFileExt(); got != fs.ExtMp4 {
		t.Fatalf("unk ext: got=%s want=%s", got, fs.ExtMp4)
	}
	if got := unk.VideoFileType(); got != fs.VideoMp4 {
		t.Fatalf("unk type: got=%v want=%v", got, fs.VideoMp4)
	}
}
