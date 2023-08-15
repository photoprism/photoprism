package meta

import (
	"github.com/photoprism/photoprism/pkg/video"
)

const CodecUnknown = ""
const CodecAv1 = string(video.CodecAV1)
const CodecAvc1 = string(video.CodecAVC)
const CodecHvc1 = string(video.CodecHEVC)
const CodecJpeg = "jpeg"
const CodecHeic = "heic"
const CodecXMP = "xmp"

// CodecAvc returns true if the video codec is AVC.
func (data Data) CodecAvc() bool {
	return data.Codec == CodecAvc1
}
