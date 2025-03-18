package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContentType(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		result := ContentType("")
		assert.Equal(t, "application/octet-stream", result)
	})
	t.Run("Mov", func(t *testing.T) {
		result := ContentType("video/quicktime")
		assert.Equal(t, "video/mp4", result)
	})
	t.Run("Mp4", func(t *testing.T) {
		result := ContentType("video/x-m4v")
		assert.Equal(t, "video/mp4", result)
	})
	t.Run("Invalid", func(t *testing.T) {
		result := ContentType("invalid")
		assert.Equal(t, "invalid", result)
	})
	t.Run("Json", func(t *testing.T) {
		result := ContentType("text/json")
		assert.Equal(t, "application/json; charset=utf-8", result)
	})
	t.Run("Html", func(t *testing.T) {
		result := ContentType("text/html")
		assert.Equal(t, "text/html; charset=utf-8", result)
	})
	t.Run("Text", func(t *testing.T) {
		result := ContentType("text/plain")
		assert.Equal(t, "text/plain; charset=utf-8", result)
	})
	t.Run("Pdf", func(t *testing.T) {
		result := ContentType("text/pdf")
		assert.Equal(t, "application/pdf", result)
	})
	t.Run("Svg", func(t *testing.T) {
		result := ContentType("image/svg")
		assert.Equal(t, "image/svg+xml", result)
	})
	t.Run("Jpeg", func(t *testing.T) {
		result := ContentType("image/jpg")
		assert.Equal(t, "image/jpeg", result)
	})
	t.Run("Mp4Avc", func(t *testing.T) {
		result := ContentType("video/mp4; codecs=\"avc1\"")
		assert.Equal(t, "video/mp4; codecs=\"avc1.4d0028\"", result)
	})
	t.Run("Mp4Avc3", func(t *testing.T) {
		result := ContentType("video/mp4; codecs=\"avc3\"")
		assert.Equal(t, "video/mp4; codecs=\"avc3.4d0028\"", result)
	})
	t.Run("Mp4Hvc", func(t *testing.T) {
		result := ContentType("video/mp4; codecs=\"hvc\"")
		assert.Equal(t, "video/mp4; codecs=\"hvc1.2.4.L153.B0\"", result)
	})
	t.Run("Mp4Hev", func(t *testing.T) {
		result := ContentType("video/mp4; codecs=\"hev\"")
		assert.Equal(t, "video/mp4; codecs=\"hev1.2.4.L153.B0\"", result)
	})
	t.Run("Webm", func(t *testing.T) {
		result := ContentType("video/webm; codecs=\"vp08\"")
		assert.Equal(t, "video/webm; codecs=\"vp8\"", result)
	})
	t.Run("WebmVp9", func(t *testing.T) {
		result := ContentType("video/webm; codecs=\"vp9\"")
		assert.Equal(t, "video/webm; codecs=\"vp09.00.10.08\"", result)
	})
	t.Run("Av1", func(t *testing.T) {
		result := ContentType("video/av1")
		assert.Equal(t, "video/mp4; codecs=\"av01.0.08H.10\"", result)
	})
	t.Run("WebmAv1", func(t *testing.T) {
		result := ContentType("video/webm; codecs=\"av1\"")
		assert.Equal(t, "video/webm; codecs=\"av01.0.08H.10\"", result)
	})
	t.Run("MkvAv1", func(t *testing.T) {
		result := ContentType("video/matroska; codecs=\"av1\"")
		assert.Equal(t, "video/matroska; codecs=\"av01.0.08H.10\"", result)
	})
	t.Run("Ogg", func(t *testing.T) {
		result := ContentType("video/ogg; codecs=\"vorbis\"")
		assert.Equal(t, "video/ogg", result)
	})
	t.Run("Ú", func(t *testing.T) {
		result := ContentType("Ú")
		assert.Equal(t, "application/octet-stream", result)
	})
}
