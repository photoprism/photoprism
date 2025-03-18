package video

import (
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media/http/header"
)

// Unknown represents an unknown and/or unsupported video format.
var Unknown = Type{
	Codec:    CodecUnknown,
	FileType: fs.TypeUnknown,
}

// Mp4 specifies the MPEG-4 Part 14 multimedia container format.
var Mp4 = Type{
	Codec:       CodecAvc1,
	FileType:    fs.VideoMp4,
	ContentType: header.ContentTypeMp4,
	WidthLimit:  8192,
	HeightLimit: 4320,
	Public:      true,
}

// Mov specifies the Apple QuickTime (QT) container format.
var Mov = Type{
	Codec:       CodecAvc1,
	FileType:    fs.VideoMov,
	ContentType: header.ContentTypeMovAvc,
	WidthLimit:  8192,
	HeightLimit: 4320,
	Public:      true,
}

// Avc specifies the MPEG-4 Advanced Video Coding (H.264) format,
// see https://en.wikipedia.org/wiki/Advanced_Video_Coding.
var Avc = Type{
	Codec:       CodecAvc1,
	FileType:    fs.VideoAvc,
	ContentType: header.ContentTypeMp4Avc,
	WidthLimit:  8192,
	HeightLimit: 4320,
	Public:      true,
}

// Hvc specifies the generally compatible High Efficiency Video Coding (H.265) format.
var Hvc = Type{
	Codec:       CodecHvc1,
	FileType:    fs.VideoHvc,
	ContentType: header.ContentTypeMp4Hvc,
	WidthLimit:  0,
	HeightLimit: 0,
	Public:      false,
}

// Hev specifies a HEVC video with parameter sets also in the Samples (not supported on macOS):
// https://ott.dolby.com/codec_test/index.html
var Hev = Type{
	Codec:       CodecHev1,
	FileType:    fs.VideoHev,
	ContentType: header.ContentTypeMp4Hev,
	WidthLimit:  0,
	HeightLimit: 0,
	Public:      false,
}

// Vvc specifies the Versatile Video Coding (H.266) format.
var Vvc = Type{
	Codec:       CodecVvc1,
	FileType:    fs.VideoVvc,
	ContentType: header.ContentTypeMp4Vvc,
	WidthLimit:  0,
	HeightLimit: 0,
	Public:      false,
}

// Evc specifies the Essential Video Coding (MPEG-5 Part 1) format.
var Evc = Type{
	Codec:       CodecEvc1,
	FileType:    fs.VideoEvc,
	ContentType: header.ContentTypeMp4Evc,
	WidthLimit:  0,
	HeightLimit: 0,
	Public:      false,
}

// Vp8 specifies a Google VP8 video in a WebM multimedia container.
var Vp8 = Type{
	Codec:       CodecVp08,
	FileType:    fs.VideoWebm,
	ContentType: header.ContentTypeWebmVp8,
	WidthLimit:  0,
	HeightLimit: 0,
	Public:      false,
}

// Vp9 specifies a Google VP9 video in a WebM multimedia container.
var Vp9 = Type{
	Codec:       CodecVp09,
	FileType:    fs.VideoWebm,
	ContentType: header.ContentTypeWebmVp9,
	WidthLimit:  0,
	HeightLimit: 0,
	Public:      false,
}

// Av1 specifies an AV1 (AOMedia Video 1) video in an MP4 multimedia container.
var Av1 = Type{
	Codec:       CodecAv01,
	FileType:    fs.VideoAv1,
	ContentType: header.ContentTypeMp4Av1Main10,
	WidthLimit:  0,
	HeightLimit: 0,
	Public:      false,
}

// WebmAv1 specifies an AV1 (AOMedia Video 1) video in a WebM multimedia container.
var WebmAv1 = Type{
	Codec:       CodecAv01,
	FileType:    fs.VideoWebm,
	ContentType: header.ContentTypeWebmAv1Main10,
	WidthLimit:  0,
	HeightLimit: 0,
	Public:      false,
}

// MkvAv1 specifies an AV1 (AOMedia Video 1) video in a Matroska multimedia container.
var MkvAv1 = Type{
	Codec:       CodecAv01,
	FileType:    fs.VideoMkv,
	ContentType: header.ContentTypeMkvAv1Main10,
	WidthLimit:  0,
	HeightLimit: 0,
	Public:      false,
}

// Theora specifies OGV video with Vorbis audio in an OGG multimedia container.
var Theora = Type{
	Codec:       CodecTheora,
	FileType:    fs.VideoTheora,
	ContentType: header.ContentTypeOgg,
	WidthLimit:  0,
	HeightLimit: 0,
	Public:      false,
}

// Webm specifies the Google WebM multimedia container format.
var Webm = Type{
	Codec:       CodecUnknown,
	FileType:    fs.VideoWebm,
	ContentType: header.ContentTypeWebm,
	WidthLimit:  0,
	HeightLimit: 0,
	Public:      false,
}
