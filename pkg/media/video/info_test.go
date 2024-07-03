package video

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
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
		info.VideoMimeType = fs.MimeTypeMP4
		info.VideoCodec = CodecAVC
		assert.Equal(t, ContentTypeAVC, info.VideoContentType())
	})
	t.Run("VideoFileExt", func(t *testing.T) {
		info := NewInfo()
		info.VideoMimeType = fs.MimeTypeMP4
		info.VideoCodec = CodecAVC
		assert.Equal(t, fs.ExtMP4, info.VideoFileExt())
	})
	t.Run("VideoFileType", func(t *testing.T) {
		info := NewInfo()
		info.VideoMimeType = fs.MimeTypeMP4
		info.VideoCodec = CodecAVC
		assert.Equal(t, fs.VideoMP4, info.VideoFileType())
	})
}
