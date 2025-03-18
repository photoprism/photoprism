package header

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContent(t *testing.T) {
	t.Run("Headers", func(t *testing.T) {
		assert.Equal(t, "Accept", Accept)
		assert.Equal(t, "Accept-Encoding", AcceptEncoding)
		assert.Equal(t, "Accept-Language", AcceptLanguage)
		assert.Equal(t, "Accept-Ranges", AcceptRanges)
		assert.Equal(t, "Content-Type", ContentType)
		assert.Equal(t, "Content-Disposition", ContentDisposition)
		assert.Equal(t, "Content-Encoding", ContentEncoding)
		assert.Equal(t, "Content-Range", ContentRange)
		assert.Equal(t, "Location", Location)
		assert.Equal(t, "Origin", Origin)
		assert.Equal(t, "Vary", Vary)
	})
	t.Run("Types", func(t *testing.T) {
		assert.Equal(t, "application/x-www-form-urlencoded", ContentTypeForm)
		assert.Equal(t, "multipart/form-data", ContentTypeMultipart)
		assert.Equal(t, "application/json", ContentTypeJson)
		assert.Equal(t, "application/json; charset=utf-8", ContentTypeJsonUtf8)
		assert.Equal(t, "text/html; charset=utf-8", ContentTypeHtml)
		assert.Equal(t, "text/plain; charset=utf-8", ContentTypeText)
		assert.Equal(t, "image/png", ContentTypePng)
		assert.Equal(t, "image/jpeg", ContentTypeJpeg)
		assert.Equal(t, "image/svg+xml", ContentTypeSVG)
		assert.Equal(t, "video/mp4; codecs=\"avc1\"", ContentTypeMp4Avc)
		assert.Equal(t, "video/mp4; codecs=\"avc1.4d0028\"", ContentTypeMp4AvcMain)
		assert.Equal(t, "video/mp4; codecs=\"avc1.640028\"", ContentTypeMp4AvcHigh)
		assert.Equal(t, "video/mp4; codecs=\"hvc1\"", ContentTypeMp4Hvc)
	})
}
