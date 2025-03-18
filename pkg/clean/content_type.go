package clean

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/media/http/header"
)

// ContentType normalizes media content type strings, see https://en.wikipedia.org/wiki/Media_type.
func ContentType(s string) string {
	if s == "" {
		return header.ContentTypeBinary
	}

	// Remove invalid characters and enforce a maximum length of 64 characters.
	s = Type(s)

	// Assume that .mov and .m4v can be played like .mp4 files,
	// since the video container formats are largely compatible.
	s = strings.Replace(s, header.ContentTypeMov, header.ContentTypeMp4, 1)
	s = strings.Replace(s, header.ContentTypeM4v, header.ContentTypeMp4, 1)

	// Normalize media type strings.
	switch s {
	case "":
		return header.ContentTypeBinary
	case
		"text/json",
		"application/json":
		return header.ContentTypeJsonUtf8
	case
		"text/htm",
		"text/html":
		return header.ContentTypeHtml
	case
		"text/plain":
		return header.ContentTypeText
	case
		"text/pdf",
		"text/x-pdf",
		"application/x-pdf",
		"application/acrobat":
		return header.ContentTypePDF
	case
		"image/svg":
		return header.ContentTypeSVG
	case
		"image/jpe",
		"image/jpg":
		return header.ContentTypeJpeg
	case
		header.ContentTypeMovAvc,
		header.ContentTypeMp4Avc,
		"video/mp4; codecs=\"avc\"":
		return header.ContentTypeMp4AvcMain // Advanced Video Coding (AVC), also known as H.264
	case
		header.ContentTypeMp4Avc3:
		return header.ContentTypeMp4Avc3Main // Advanced Video Coding (AVC) Bitstream
	case
		header.ContentTypeMp4Hvc,
		"video/mp4; codecs=\"hvc\"",
		"video/mp4; codecs=\"hevc\"":
		return header.ContentTypeMp4HvcMain10 // HEVC Mp4 Main10 Profile
	case
		header.ContentTypeMp4Hev,
		"video/mp4; codecs=\"hev\"":
		return header.ContentTypeMp4HevMain10 // HEVC with parameter sets also in the Samples, not supported on macOS
	case
		"video/webm; codecs=\"vp08\"":
		return header.ContentTypeWebmVp8 // Google WebM container with VP8 video
	case
		header.ContentTypeWebmVp9,
		"video/webm; codecs=\"vp9\"":
		return header.ContentTypeWebmVp9Main // Google WebM container with VP9 video
	case
		header.ContentTypeMp4Av1,
		"video/mp4; codecs=\"av1\"",
		header.ContentTypeAv1,
		"video/av1",
		"video/av1; codecs=\"av01\"",
		"video/AV1; codecs=\"av01\"":
		return header.ContentTypeMp4Av1Main10 // MP4 container with AV1 video
	case
		header.ContentTypeWebmAv1,
		"video/webm; codecs=\"av1\"",
		"video/webm; codecs=\"av1c\"",
		"video/webm; codecs=\"av1C\"":
		return header.ContentTypeWebmAv1Main10 // Google WebM container with AV1 video
	case
		header.ContentTypeMkvAv1,
		"video/matroska; codecs=\"av1\"",
		"video/matroska; codecs=\"av1c\"",
		"video/matroska; codecs=\"av1C\"":
		return header.ContentTypeMkvAv1Main10 // Matroska container with AV1 video
	case
		header.ContentTypeOggVorbis,
		header.ContentTypeOggTheora:
		return header.ContentTypeOgg
	}

	return s
}
