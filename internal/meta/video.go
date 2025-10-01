package meta

import (
	"github.com/photoprism/photoprism/pkg/media/video"
)

// Common codec labels encountered in metadata.
const (
	CodecUnknown = ""
	CodecJpeg    = "jpeg"
	CodecHeic    = "heic"
	CodecXMP     = "xmp"
)

// CodecAvc returns true if the video codec is AVC.
func (data Data) CodecAvc() bool {
	switch data.Codec {
	case video.CodecAvc1, video.CodecAvc2, video.CodecAvc3, video.CodecAvc4:
		return true
	default:
		return false
	}
}
