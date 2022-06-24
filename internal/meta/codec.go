package meta

import (
	"github.com/photoprism/photoprism/pkg/video"
)

const CodecUnknown = ""
const CodecAv1 = string(video.CodecAV1)
const CodecVP9 = string(video.CodecVP9)
const CodecAvc1 = string(video.CodecAVC)
const CodecJpeg = "jpeg"
const CodecHeic = "heic"
const CodecXMP = "xmp"

// CodecAvc returns true if the video format is MPEG-4 AVC.
func (data Data) CodecAvc() bool {
	return data.Codec == CodecAvc1
}
