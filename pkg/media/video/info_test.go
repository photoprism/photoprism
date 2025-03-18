package video

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media/http/header"
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
