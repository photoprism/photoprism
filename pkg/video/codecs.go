package video

type Codec string

// Check browser support: https://cconcolato.github.io/media-mime-support/

const (
	UnknownCodec Codec = ""
	CodecAVC     Codec = "avc1"
	CodecHEVC    Codec = "hvc1"
	CodecVVC     Codec = "vvc"
	CodecAV1     Codec = "av01"
	CodecVP8     Codec = "vp8"
	CodecVP9     Codec = "vp9"
	CodecOGV     Codec = "ogv"
	CodecWebM    Codec = "webm"
)

// Codecs maps identifiers to codecs.
var Codecs = StandardCodecs{
	"":         UnknownCodec,
	"a_opus":   UnknownCodec,
	"a_vorbis": UnknownCodec,
	"avc":      CodecAVC,
	"avc1":     CodecAVC,
	"v_avc":    CodecAVC,
	"v_avc1":   CodecAVC,
	"hevc":     CodecHEVC,
	"hvc":      CodecHEVC,
	"hvc1":     CodecHEVC,
	"v_hvc":    CodecHEVC,
	"v_hvc1":   CodecHEVC,
	"vvc":      CodecVVC,
	"v_vvc":    CodecVVC,
	"av1":      CodecAV1,
	"av01":     CodecAV1,
	"v_av1":    CodecAV1,
	"v_av01":   CodecAV1,
	"vp8":      CodecVP8,
	"vp80":     CodecVP8,
	"v_vp8":    CodecVP8,
	"vp9":      CodecVP9,
	"vp90":     CodecVP9,
	"v_vp9":    CodecVP9,
	"ogv":      CodecOGV,
	"webm":     CodecWebM,
}

// StandardCodecs maps names to known codecs.
type StandardCodecs map[string]Codec
