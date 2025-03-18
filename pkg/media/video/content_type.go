package video

import (
	"fmt"
	"mime"
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media/http/header"
)

// ContentType returns a normalized video content type strings based on the video file type and codec.
func ContentType(mediaType, fileType, videoCodec string, hdr bool) string {
	if mediaType == "" && fileType == "" && videoCodec == "" {
		return header.ContentTypeBinary
	}

	if mediaType == "" {
		videoCodec = Codecs[videoCodec]

		if fs.VideoMov.Equal(fileType) || fs.VideoM4v.Equal(fileType) {
			fileType = fs.VideoMp4.String()
		}

		switch {
		case fs.VideoMp4.Equal(fileType) && videoCodec == CodecAvc3:
			if hdr {
				mediaType = header.ContentTypeMp4Avc3High10 // AVC, High Profile 10-bit HDR
			} else {
				mediaType = header.ContentTypeMp4Avc3Main // AVC, Main Profile
			}
		case fs.VideoAvc.Equal(fileType) || fs.VideoMp4.Equal(fileType) && videoCodec == CodecAvc1:
			if hdr {
				mediaType = header.ContentTypeMp4AvcHigh10 // AVC, High Profile 10-bit HDR
			} else {
				mediaType = header.ContentTypeMp4AvcMain // AVC, Main Profile
			}
		case fs.VideoHvc.Equal(fileType) || fs.VideoMp4.Equal(fileType) && videoCodec == CodecHvc1:
			mediaType = header.ContentTypeMp4HvcMain10 // HEVC Main Profile, 10-Bit HDR
		case fs.VideoHev.Equal(fileType) || fs.VideoMp4.Equal(fileType) && videoCodec == CodecHev1:
			mediaType = header.ContentTypeMp4HevMain10 // HEVC Main Profile, 10-Bit HDR
		case fs.VideoVvc.Equal(fileType) || fs.VideoMp4.Equal(fileType) && videoCodec == CodecVvc1:
			mediaType = header.ContentTypeMp4Vvc // Versatile Video Coding (VVC)
		case fs.VideoEvc.Equal(fileType) || fs.VideoMp4.Equal(fileType) && videoCodec == CodecEvc1:
			mediaType = header.ContentTypeMp4Evc // MPEG-5 Essential Video Coding (EVC)
		case fs.VideoVp8.Equal(fileType) || videoCodec == CodecVp08:
			mediaType = header.ContentTypeWebmVp8
		case fs.VideoVp9.Equal(fileType) || videoCodec == CodecVp09:
			mediaType = header.ContentTypeWebmVp9
		case fs.VideoMp4.Equal(fileType) && videoCodec == CodecAv01:
			mediaType = header.ContentTypeMp4Av1Main10
		case fs.VideoWebm.Equal(fileType) && videoCodec == CodecAv01:
			mediaType = header.ContentTypeWebmAv1Main10
		case fs.VideoMkv.Equal(fileType) && videoCodec == CodecAv01:
			mediaType = header.ContentTypeMkvAv1Main10
		case fs.VideoAv1.Equal(fileType):
			mediaType = header.ContentTypeAv1
		case fs.VideoTheora.Equal(fileType) || videoCodec == CodecTheora:
			mediaType = header.ContentTypeOgg
		case fs.VideoWebm.Equal(fileType):
			mediaType = header.ContentTypeWebm
		case fs.VideoMp4.Equal(fileType):
			mediaType = header.ContentTypeMp4
		case fs.VideoMkv.Equal(fileType):
			mediaType = header.ContentTypeMkv
		}
	}

	// Add codec parameter, if possible.
	if mediaType != "" && !strings.Contains(mediaType, ";") {
		if codec, found := Codecs[videoCodec]; found && codec != "" {
			mediaType = fmt.Sprintf("%s; codecs=\"%s\"", mediaType, codec)
		}
	}

	// Normalize the media content type string.
	mediaType = clean.ContentType(mediaType)

	// Adjust codec details for HDR video content.
	if hdr {
		switch mediaType {
		case
			header.ContentTypeMovAvc,
			header.ContentTypeMovAvcMain,
			header.ContentTypeMovAvcHigh,
			header.ContentTypeMp4Avc,
			header.ContentTypeMp4AvcBaseline,
			header.ContentTypeMp4AvcMain,
			header.ContentTypeMp4AvcHigh:
			if Codecs[videoCodec] == CodecAvc3 {
				mediaType = header.ContentTypeMp4Avc3High10
			} else {
				mediaType = header.ContentTypeMp4AvcHigh10
			}
		case
			header.ContentTypeMp4Avc3,
			header.ContentTypeMp4Avc3Main,
			header.ContentTypeMp4Avc3High:
			mediaType = header.ContentTypeMp4Avc3High10
		case
			header.ContentTypeMp4Hvc,
			header.ContentTypeMovHvc,
			header.ContentTypeMp4HvcMain:
			mediaType = header.ContentTypeMp4HvcMain10
		case
			header.ContentTypeMp4Hev,
			header.ContentTypeMp4HevMain:
			mediaType = header.ContentTypeMp4HevMain10
		case
			header.ContentTypeAv1,
			header.ContentTypeMovAv1,
			header.ContentTypeMp4Av1,
			header.ContentTypeMp4Av1Main:
			mediaType = header.ContentTypeMp4Av1Main10
		case
			header.ContentTypeWebmAv1,
			header.ContentTypeWebmAv1Main:
			mediaType = header.ContentTypeWebmAv1Main10
		case
			header.ContentTypeMkvAv1,
			header.ContentTypeMkvAv1Main:
			mediaType = header.ContentTypeMkvAv1Main10
		}
	}

	return mediaType
}

// Compatible tests if the video content types are expected to be compatible,
func Compatible(contentType1, contentType2 string) bool {
	// Content is likely compatible if the content type strings match exactly (case-insensitive).
	if contentType1 == "" || contentType2 == "" {
		return false
	} else if strings.EqualFold(contentType1, contentType2) {
		return true
	}

	// Sanitize and normalize content type strings.
	contentType1 = clean.ContentType(contentType1)
	contentType2 = clean.ContentType(contentType2)

	// Parse content type strings.
	mediaType1, params1, err1 := mime.ParseMediaType(contentType1)
	mediaType2, params2, err2 := mime.ParseMediaType(contentType2)

	// If parsing fails, assume the content is invalid or incompatible.
	if err1 != nil || err2 != nil {
		return false
	} else if len(params1) == 0 && len(params2) == 0 {
		return strings.EqualFold(mediaType1, mediaType2)
	}

	// If the media types don't match, assume the content is incompatible.
	if !strings.EqualFold(mediaType1, mediaType2) {
		return false
	}

	// Compare the media codecs.
	codec1 := params1["codecs"]
	codec2 := params2["codecs"]

	// Content is likely compatible if the full codec details match (case-insensitive).
	if strings.EqualFold(codec1, codec2) {
		return true
	}

	// Compare main codec names.
	codec1, _, _ = strings.Cut(codec1, ",")
	codec2, _, _ = strings.Cut(codec2, ",")

	codecName1, _, _ := strings.Cut(strings.TrimSpace(codec1), ".")
	codecName2, _, _ := strings.Cut(strings.TrimSpace(codec2), ".")

	// Content is likely compatible if the name of the main codec matches (case-insensitive).
	return strings.EqualFold(codecName1, codecName2)
}
