package video

type Codec string

// String returns the codec name as string.
func (c Codec) String() string {
	return string(c)
}

// Check browser support: https://cconcolato.github.io/media-mime-support/

const (
	CodecUnknown Codec = ""
	CodecAVC     Codec = "avc1"
	CodecHVC     Codec = "hvc1"
	CodecVVC     Codec = "vvc"
	CodecAV1     Codec = "av01"
	CodecVP8     Codec = "vp8"
	CodecVP9     Codec = "vp9"
	CodecOGV     Codec = "ogv"
	CodecWebM    Codec = "webm"
)

// Codecs maps identifiers to codecs.
var Codecs = StandardCodecs{
	"":         CodecUnknown,
	"a_opus":   CodecUnknown,
	"a_vorbis": CodecUnknown,
	"avc":      CodecAVC,
	"avc1":     CodecAVC,
	"v_avc":    CodecAVC,
	"v_avc1":   CodecAVC,
	"hevc":     CodecHVC,
	"hvc":      CodecHVC,
	"hvc1":     CodecHVC,
	"v_hvc":    CodecHVC,
	"v_hvc1":   CodecHVC,
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
