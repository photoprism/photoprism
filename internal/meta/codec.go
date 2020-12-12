package meta

const CodecUnknown = ""
const CodecJpeg = "jpeg"
const CodecAvc1 = "avc1"
const CodecHeic = "heic"
const CodecXMP = "xmp"

// CodecAvc returns true if the video format is MPEG-4 AVC.
func (data Data) CodecAvc() bool {
	return data.Codec == CodecAvc1
}
