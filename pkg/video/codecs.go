package video

type Codec string

const (
	UnknownCodec Codec = ""
	CodecAVC     Codec = "avc1"
	CodecHEVC    Codec = "hvc1"
	CodecVVC     Codec = "vvc"
	CodecAV1     Codec = "av01"
)

// Codecs maps identifiers to codecs.
var Codecs = StandardCodecs{
	"":     UnknownCodec,
	"avc":  CodecAVC,
	"avc1": CodecAVC,
	"hvc1": CodecHEVC,
	"hvc":  CodecHEVC,
	"hevc": CodecHEVC,
	"vvc":  CodecVVC,
	"av1":  CodecAV1,
	"av01": CodecAV1,
}

// StandardCodecs maps names to known codecs.
type StandardCodecs map[string]Codec
