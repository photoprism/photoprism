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

// AV1 aka AOMedia Video 1.
var AV1 = Type{
	File:   fs.VideoAV1,
	Codec:  CodecAV1,
	Width:  0,
	Height: 0,
	Public: false,
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
