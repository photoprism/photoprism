package video

import (
	"github.com/photoprism/photoprism/pkg/fs"
)

// Unknown represents an unknown video file type.
var Unknown = Type{
	Codec:    CodecUnknown,
	FileType: fs.TypeUnknown,
}

// MP4 is a Multimedia Container (MPEG-4 Part 14).
var MP4 = Type{
	Codec:       CodecAVC,
	FileType:    fs.VideoMP4,
	WidthLimit:  8192,
	HeightLimit: 4320,
	Public:      true,
}

// MOV are QuickTime videos based on the MPEG-4 format,
var MOV = Type{
	Codec:       CodecAVC,
	FileType:    fs.VideoMOV,
	WidthLimit:  8192,
	HeightLimit: 4320,
	Public:      true,
}

// AVC aka Advanced Video Coding (H.264).
var AVC = Type{
	Codec:       CodecAVC,
	FileType:    fs.VideoAVC,
	WidthLimit:  8192,
	HeightLimit: 4320,
	Public:      true,
}

// HEVC aka High Efficiency Video Coding (H.265).
var HEVC = Type{
	Codec:       CodecHVC,
	FileType:    fs.VideoHEVC,
	WidthLimit:  0,
	HeightLimit: 0,
	Public:      false,
}

// EVC aka Essential Video Coding (MPEG-5 Part 1).
var EVC = Type{
	Codec:       CodecEVC,
	FileType:    fs.VideoEVC,
	WidthLimit:  0,
	HeightLimit: 0,
	Public:      false,
}

// VVC aka Versatile Video Coding (H.266).
var VVC = Type{
	Codec:       CodecVVC,
	FileType:    fs.VideoVVC,
	WidthLimit:  0,
	HeightLimit: 0,
	Public:      false,
}

// VP8 + Google WebM.
var VP8 = Type{
	Codec:       CodecVP8,
	FileType:    fs.VideoWebM,
	WidthLimit:  0,
	HeightLimit: 0,
	Public:      false,
}

// VP9 + Google WebM.
var VP9 = Type{
	Codec:       CodecVP9,
	FileType:    fs.VideoWebM,
	WidthLimit:  0,
	HeightLimit: 0,
	Public:      false,
}

// AV1 + Google WebM.
var AV1 = Type{
	Codec:       CodecAV1,
	FileType:    fs.VideoWebM,
	WidthLimit:  0,
	HeightLimit: 0,
	Public:      false,
}

// OGV aka Ogg/Theora.
var OGV = Type{
	Codec:       CodecOGV,
	FileType:    fs.VideoOGV,
	WidthLimit:  0,
	HeightLimit: 0,
	Public:      false,
}

// WebM Container.
var WebM = Type{
	Codec:       CodecUnknown,
	FileType:    fs.VideoWebM,
	WidthLimit:  0,
	HeightLimit: 0,
	Public:      false,
}
