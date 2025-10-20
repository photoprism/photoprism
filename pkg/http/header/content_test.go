package header

import (
	"net/http"
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
		assert.Equal(t, "Content-Language", ContentLanguage)
		assert.Equal(t, "Content-Length", ContentLength)
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
		assert.Equal(t, "application/javascript", ContentTypeJavaScript)
		assert.Equal(t, "text/css", ContentTypeCSS)
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

func TestHasContentType(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		assert.True(t, HasContentType(&http.Header{"Content-Type": []string{"multipart/form-data"}}, ContentTypeMultipart))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, HasContentType(nil, ContentTypeMultipart))
		assert.False(t, HasContentType(&http.Header{"Content-Type": []string{"application/json"}}, ContentTypeMultipart))
	})
}
