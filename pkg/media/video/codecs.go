package video

type Codec = string

// Standard video Codec types, see:
// - https://mp4ra.org/registered-types/codecs
// - https://developer.mozilla.org/en-US/docs/Web/Media/Guides/Formats/Video_codecs
//
// Browser support can be tested by visiting one of the following sites:
// - https://ott.dolby.com/codec_test/index.html
// - https://dmnsgn.github.io/media-codecs/
// - https://cconcolato.github.io/media-mime-support/
// - https://thorium.rocks/misc/h265-tester.html
const (
	CodecUnknown Codec = ""
	CodecAvc1    Codec = "avc1" // Advanced Video Coding (AVC), also known as H.264
	CodecDva1    Codec = "dva1" // AVC-based Dolby Vision derived from avc1
	CodecAvc2    Codec = "avc2" // Advanced Video Coding (AVC), might not be fully supported on macOS
	CodecAvc3    Codec = "avc3" // Advanced Video Coding (AVC), might not be fully supported on macOS
	CodecAvc4    Codec = "avc4" // Advanced Video Coding (AVC), might not be fully supported on macOS
	CodecDvav    Codec = "dvav" // AVC-based Dolby Vision derived from avc3
	CodecHvc1    Codec = "hvc1" // High Efficiency Video Coding (HEVC), also known as H.265
	CodecDvh1    Codec = "dvh1" // HEVC-based Dolby Vision derived from hvc1
	CodecHvc2    Codec = "hvc2" // HEVC video with constrained extractors and/or aggregators and parameter sets only in the Sample Entry
	CodecHvc3    Codec = "hvc3" // HEVC video with extractors and/or aggregators and parameter sets only in the Sample Entry
	CodecHev1    Codec = "hev1" // HEVC video with parameter sets also in the Samples, not supported on macOS
	CodecDvhe    Codec = "dvhe" // HEVC-based Dolby Vision derived from hev1
	CodecHev2    Codec = "hev2" // HEVC video with constrained extractors and/or aggregators and parameter sets in the Sample Entry or samples
	CodecHev3    Codec = "hev3" // HEVC video with extractors and/or aggregators and parameter sets in the Sample Entry or samples
	CodecVvc1    Codec = "vvc1" // Versatile Video Coding (VVC), also known as H.266
	CodecEvc1    Codec = "evc1" // MPEG-5 Essential Video Coding (EVC), also known as ISO/IEC 23094-1
	CodecAv01    Codec = "av01" // AOMedia Video 1 (AV1)
	CodecVp08    Codec = "vp08" // Google VP8
	CodecVp09    Codec = "vp09" // Google VP9
	CodecTheora  Codec = "ogv"  // Ogg Vorbis Video
	CodecWebm    Codec = "webm" // Google WebM
)

// Codecs maps supported string identifiers to standard Codec types.
//
// Complete list of registered codecs and formats:
// https://mp4ra.org/registered-types/codecs
var Codecs = StandardCodecs{
	"":                 CodecUnknown,
	"a_opus":           CodecUnknown,
	"a_vorbis":         CodecUnknown,
	CodecAvc1:          CodecAvc1,
	CodecDva1:          CodecAvc1,
	CodecAvc2:          CodecAvc1,
	CodecAvc4:          CodecAvc1,
	"avc":              CodecAvc1,
	"v_avc":            CodecAvc1,
	"v_avc1":           CodecAvc1,
	"v_avc2":           CodecAvc1,
	"v_avc4":           CodecAvc1,
	"iso/avc":          CodecAvc1,
	"iso/avc1":         CodecAvc1,
	"iso/avc2":         CodecAvc1,
	"iso/avc4":         CodecAvc1,
	"v_mpeg4/avc":      CodecAvc1,
	"v_mpeg4/avc1":     CodecAvc1,
	"v_mpeg4/iso/avc":  CodecAvc1,
	"v_mpeg4/iso/avc1": CodecAvc1,
	CodecAvc3:          CodecAvc3,
	CodecDvav:          CodecAvc3,
	"v_avc3":           CodecAvc3,
	CodecHvc1:          CodecHvc1,
	CodecDvh1:          CodecHvc1,
	CodecHvc2:          CodecHvc1,
	CodecHvc3:          CodecHvc1,
	"hevc":             CodecHvc1,
	"v_hevc":           CodecHvc1,
	"hevC":             CodecHvc1,
	"hvc":              CodecHvc1,
	"v_hvc":            CodecHvc1,
	"v_hvc1":           CodecHvc1,
	"hvcC":             CodecHvc1,
	"hvcc":             CodecHvc1,
	CodecHev1:          CodecHev1,
	CodecDvhe:          CodecHev1,
	CodecHev2:          CodecHev1,
	CodecHev3:          CodecHev1,
	"hev":              CodecHev1,
	"v_hev":            CodecHev1,
	"v_hev1":           CodecHev1,
	CodecVvc1:          CodecVvc1,
	"vvc":              CodecVvc1,
	"vvcC":             CodecVvc1,
	"vvcc":             CodecVvc1,
	"v_vvc":            CodecVvc1,
	"v_vvc1":           CodecVvc1,
	CodecEvc1:          CodecEvc1,
	"evc":              CodecEvc1,
	"evcC":             CodecEvc1,
	"evcc":             CodecEvc1,
	"v_evc":            CodecEvc1,
	"v_evc1":           CodecEvc1,
	CodecAv01:          CodecAv01,
	"av1f":             CodecAv01,
	"av1m":             CodecAv01,
	"av1M":             CodecAv01,
	"av1s":             CodecAv01,
	"av1c":             CodecAv01,
	"av1C":             CodecAv01,
	"av1":              CodecAv01,
	"v_av1":            CodecAv01,
	"v_av01":           CodecAv01,
	CodecVp08:          CodecVp08,
	"vp8":              CodecVp08,
	"vp80":             CodecVp08,
	"v_vp8":            CodecVp08,
	"v_vp08":           CodecVp08,
	CodecVp09:          CodecVp09,
	"vp9":              CodecVp09,
	"vp90":             CodecVp09,
	"v_vp9":            CodecVp09,
	"v_vp09":           CodecVp09,
	CodecTheora:        CodecTheora,
	"v_ogv":            CodecTheora,
	"theora":           CodecTheora,
	"v_theora":         CodecTheora,
	CodecWebm:          CodecWebm,
}

// StandardCodecs maps strings to codec types.
type StandardCodecs map[string]Codec
