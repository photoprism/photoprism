package meta

const CodecUnknown = ""
const CodecJpeg = "jpeg"
const CodecAvc1 = "avc1"

// CodecAvc1 returns true if the video is encoded with H.264/AVC
func (data Data) CodecAvc1() bool {
	return data.Codec == CodecAvc1
}
