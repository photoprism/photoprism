package video

import (
	"github.com/photoprism/photoprism/pkg/fs"
)

// MP4 is a Multimedia Container (MPEG-4 Part 14).
var MP4 = Type{
	File:   fs.VideoMP4,
	Codec:  CodecAVC,
	Width:  0,
	Height: 0,
	Public: true,
}

// AVC aka Advanced Video Coding (H.264).
var AVC = Type{
	File:   fs.VideoAVC,
	Codec:  CodecAVC,
	Width:  0,
	Height: 0,
	Public: true,
}

// HEVC aka High Efficiency Video Coding (H.265).
var HEVC = Type{
	File:   fs.VideoHEVC,
	Codec:  CodecHEVC,
	Width:  0,
	Height: 0,
	Public: false,
}

// VVC aka Versatile Video Coding (H.266).
var VVC = Type{
	File:   fs.VideoVVC,
	Codec:  CodecVVC,
	Width:  0,
	Height: 0,
	Public: false,
}

// VP8 + Google WebM.
var VP8 = Type{
	File:   fs.VideoWebM,
	Codec:  CodecVP8,
	Width:  0,
	Height: 0,
	Public: false,
}

// VP9 + Google WebM.
var VP9 = Type{
	File:   fs.VideoWebM,
	Codec:  CodecVP9,
	Width:  0,
	Height: 0,
	Public: false,
}

// AV1 + Google WebM.
var AV1 = Type{
	File:   fs.VideoWebM,
	Codec:  CodecAV1,
	Width:  0,
	Height: 0,
	Public: false,
}

// OGV aka Ogg/Theora.
var OGV = Type{
	File:   fs.VideoOGV,
	Codec:  CodecOGV,
	Width:  0,
	Height: 0,
	Public: false,
}

// WebM Container.
var WebM = Type{
	File:   fs.VideoWebM,
	Codec:  UnknownCodec,
	Width:  0,
	Height: 0,
	Public: false,
}
