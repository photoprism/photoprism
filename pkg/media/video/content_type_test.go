package video

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/media/http/header"
)

func TestContentType(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "application/octet-stream", ContentType("", "", "", false))
	})
	t.Run("Mkv", func(t *testing.T) {
		assert.Equal(t, header.ContentTypeMkv, ContentType("", "mkv", "", false))
		assert.Equal(t, header.ContentTypeMkv, ContentType(header.ContentTypeMkv, "", "", false))
	})
	t.Run("Mov", func(t *testing.T) {
		assert.Equal(t, header.ContentTypeMp4, ContentType(header.ContentTypeMov, "", "", false))
	})
	t.Run("M4v", func(t *testing.T) {
		assert.Equal(t, header.ContentTypeMp4, ContentType("", "m4v", "", false))
		assert.Equal(t, header.ContentTypeMp4, ContentType(header.ContentTypeM4v, "", "", false))
		assert.Equal(t, header.ContentTypeMp4AvcMain, ContentType(header.ContentTypeM4v, "", CodecAvc1, false))
		assert.Equal(t, header.ContentTypeMp4AvcHigh10, ContentType(header.ContentTypeM4v, "", CodecAvc1, true))
		assert.Equal(t, header.ContentTypeMp4HvcMain10, ContentType(header.ContentTypeM4v, "", CodecHvc1, false))
		assert.Equal(t, header.ContentTypeMp4HvcMain10, ContentType(header.ContentTypeM4v, "", CodecHvc1, true))
	})
	t.Run("Mov/Hvc", func(t *testing.T) {
		assert.Equal(t, header.ContentTypeMp4HvcMain10, ContentType(header.ContentTypeMov, "mov", CodecHvc1, false))
	})
	t.Run("Mp4", func(t *testing.T) {
		assert.Equal(t, header.ContentTypeMp4, ContentType(header.ContentTypeMp4, "", "", false))
		assert.Equal(t, header.ContentTypeMp4, ContentType(header.ContentTypeMp4, "", "", true))
	})
	t.Run("Mp4/Avc", func(t *testing.T) {
		assert.Equal(t, header.ContentTypeMp4AvcMain, ContentType(header.ContentTypeMp4, "", CodecAvc1, false))
		assert.Equal(t, header.ContentTypeMp4AvcHigh10, ContentType(header.ContentTypeMp4, "", CodecAvc1, true))
		assert.Equal(t, header.ContentTypeMp4AvcMain, ContentType("", "mp4", CodecAvc1, false))
		assert.Equal(t, header.ContentTypeMp4AvcHigh10, ContentType("", "mp4", CodecAvc1, true))
	})
	t.Run("Mp4/Avc3", func(t *testing.T) {
		assert.Equal(t, header.ContentTypeMp4Avc3Main, ContentType(header.ContentTypeMp4, "", CodecAvc3, false))
		assert.Equal(t, header.ContentTypeMp4Avc3High10, ContentType(header.ContentTypeMp4, "", CodecAvc3, true))
		assert.Equal(t, header.ContentTypeMp4Avc3Main, ContentType("", "mp4", CodecAvc3, false))
		assert.Equal(t, header.ContentTypeMp4Avc3High10, ContentType("", "mp4", CodecAvc3, true))
	})
	t.Run("Mp4/Hvc", func(t *testing.T) {
		assert.Equal(t, header.ContentTypeMp4HvcMain10, ContentType(header.ContentTypeMp4, "", CodecHvc1, false))
		assert.Equal(t, header.ContentTypeMp4HvcMain10, ContentType(header.ContentTypeMp4, "", CodecHvc1, true))
		assert.Equal(t, header.ContentTypeMp4HvcMain10, ContentType("", "mp4", CodecHvc1, false))
		assert.Equal(t, header.ContentTypeMp4HvcMain10, ContentType("", "mp4", CodecHvc1, true))
	})
	t.Run("Mp4/Hev", func(t *testing.T) {
		assert.Equal(t, header.ContentTypeMp4HevMain10, ContentType(header.ContentTypeMp4, "", CodecHev1, false))
		assert.Equal(t, header.ContentTypeMp4HevMain10, ContentType(header.ContentTypeMp4, "", CodecHev1, true))
		assert.Equal(t, header.ContentTypeMp4HevMain10, ContentType("", "mp4", CodecHev1, false))
		assert.Equal(t, header.ContentTypeMp4HevMain10, ContentType("", "mp4", CodecHev1, true))
	})
	t.Run("Mp4/Vvc", func(t *testing.T) {
		assert.Equal(t, header.ContentTypeMp4Vvc, ContentType(header.ContentTypeMp4, "", CodecVvc1, false))
		assert.Equal(t, header.ContentTypeMp4Vvc, ContentType(header.ContentTypeMp4, "", CodecVvc1, true))
		assert.Equal(t, header.ContentTypeMp4Vvc, ContentType("", "vvc", CodecVvc1, false))
		assert.Equal(t, header.ContentTypeMp4Vvc, ContentType("", "vvc", CodecVvc1, true))
	})
	t.Run("Mp4/Evc", func(t *testing.T) {
		assert.Equal(t, header.ContentTypeMp4Evc, ContentType(header.ContentTypeMp4, "", CodecEvc1, false))
		assert.Equal(t, header.ContentTypeMp4Evc, ContentType(header.ContentTypeMp4, "", CodecEvc1, true))
		assert.Equal(t, header.ContentTypeMp4Evc, ContentType("", "evc", CodecEvc1, false))
		assert.Equal(t, header.ContentTypeMp4Evc, ContentType("", "evc", CodecEvc1, true))
	})
	t.Run("VP8", func(t *testing.T) {
		assert.Equal(t, header.ContentTypeWebmVp8, ContentType("", "", CodecVp08, false))
		assert.Equal(t, header.ContentTypeWebmVp8, ContentType("", "", CodecVp08, true))
		assert.Equal(t, header.ContentTypeWebmVp8, ContentType("", "vp8", "", false))
		assert.Equal(t, header.ContentTypeWebmVp8, ContentType("", "vp8", "", true))
	})
	t.Run("VP9", func(t *testing.T) {
		assert.Equal(t, header.ContentTypeWebmVp9Main, ContentType("", "", CodecVp09, false))
		assert.Equal(t, header.ContentTypeWebmVp9Main, ContentType("", "", CodecVp09, true))
		assert.Equal(t, header.ContentTypeWebmVp9Main, ContentType("", "vp9", "", false))
		assert.Equal(t, header.ContentTypeWebmVp9Main, ContentType("", "vp9", "", true))
	})
	t.Run("Webmm/Av1", func(t *testing.T) {
		assert.Equal(t, header.ContentTypeWebmAv1Main10, ContentType("", "webm", CodecAv01, false))
		assert.Equal(t, header.ContentTypeWebmAv1Main10, ContentType("", "webm", CodecAv01, true))
	})
	t.Run("Mp4/Av1", func(t *testing.T) {
		assert.Equal(t, header.ContentTypeMp4Av1Main10, ContentType(header.ContentTypeMp4, "mp4", CodecAv01, false))
		assert.Equal(t, header.ContentTypeMp4Av1Main10, ContentType(header.ContentTypeMp4, "mp4", CodecAv01, true))
		assert.Equal(t, header.ContentTypeMp4Av1Main10, ContentType("", "mp4", CodecAv01, false))
		assert.Equal(t, header.ContentTypeMp4Av1Main10, ContentType("", "mp4", CodecAv01, true))
	})
	t.Run("Mkv/Av1", func(t *testing.T) {
		assert.Equal(t, header.ContentTypeMkvAv1Main10, ContentType("", "mkv", CodecAv01, false))
		assert.Equal(t, header.ContentTypeMkvAv1Main10, ContentType("", "mkv", CodecAv01, true))
	})
	t.Run("Av1", func(t *testing.T) {
		assert.Equal(t, header.ContentTypeMp4Av1Main10, ContentType("", "av1", "", false))
		assert.Equal(t, header.ContentTypeMp4Av1Main10, ContentType("", "av1", "", true))
	})
	t.Run("Ogg", func(t *testing.T) {
		assert.Equal(t, header.ContentTypeOgg, ContentType("", "ogv", "", false))
		assert.Equal(t, header.ContentTypeOgg, ContentType("", "ogv", "", true))
	})
	t.Run("Webm", func(t *testing.T) {
		assert.Equal(t, header.ContentTypeWebm, ContentType("", "webm", "", false))
		assert.Equal(t, header.ContentTypeWebm, ContentType("", "webm", "", true))
	})
}

func TestCompatible(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		assert.True(t, Compatible(header.ContentTypeAv1, "video/av1"))
		assert.True(t, Compatible(header.ContentTypeAv1, "Video/Av1"))
		assert.True(t, Compatible(header.ContentTypeJpeg, header.ContentTypeJpeg))
		assert.True(t, Compatible(header.ContentTypeMp4, header.ContentTypeMov))
		assert.True(t, Compatible(header.ContentTypeMp4Hvc, header.ContentTypeMp4HvcMain))
		assert.True(t, Compatible(header.ContentTypeMp4Hvc, header.ContentTypeMp4HvcMain10))
		assert.True(t, Compatible(header.ContentTypeMp4Avc, header.ContentTypeMp4AvcHigh))
		assert.True(t, Compatible(header.ContentTypeMp4Avc, header.ContentTypeMp4AvcHigh10))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, Compatible("", ""))
		assert.False(t, Compatible("", header.ContentTypeMov))
		assert.False(t, Compatible(header.ContentTypeMp4, header.ContentTypeMp4Avc))
		assert.False(t, Compatible(header.ContentTypeMp4Avc, header.ContentTypeMp4Hvc))
		assert.False(t, Compatible(header.ContentTypeWebm, header.ContentTypeMkv))
		assert.False(t, Compatible(header.ContentTypeWebmAv1, header.ContentTypeMkvAv1))
		assert.False(t, Compatible(header.ContentTypeWebmAv1Main10, header.ContentTypeMkvAv1Main10))
	})
}
