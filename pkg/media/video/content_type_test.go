package video

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestContentType(t *testing.T) {
	t.Run("QuickTime", func(t *testing.T) {
		assert.Equal(t, fs.MimeTypeMOV, ContentType(fs.MimeTypeMOV, ""))
	})
	t.Run("QuickTime_HVC", func(t *testing.T) {
		assert.Equal(t, `video/quicktime; codecs="hvc1"`, ContentType(fs.MimeTypeMOV, CodecHVC))
	})
	t.Run("MP4", func(t *testing.T) {
		assert.Equal(t, fs.MimeTypeMP4, ContentType(fs.MimeTypeMP4, ""))
	})
	t.Run("MP4_AVC", func(t *testing.T) {
		assert.Equal(t, ContentTypeAVC, ContentType(fs.MimeTypeMP4, CodecAVC))
	})
	t.Run("MP4_HVC", func(t *testing.T) {
		assert.Equal(t, `video/mp4; codecs="hvc1"`, ContentType(fs.MimeTypeMP4, CodecHVC))
	})
}
