package video

// Types maps identifiers to standards, see https://tools.woolyss.com/html5-canplaytype-tester/:
var Types = Standards{
	"":          Avc, // ↓ Default format: Advanced Video Coding (AVC) in MP4 container
	"h264":      Avc,
	"h.264":     Avc,
	"avc":       Avc,
	"mp4_avc":   Avc,
	"mp4-avc":   Avc,
	"avc1":      Avc,
	"avc2":      Avc,
	"avc3":      Avc,
	"dva1":      Avc,
	"dvav":      Avc,
	"h265":      Hvc, // ↓ High Efficiency Video Coding (HEVC) in MP4 container
	"h.265":     Hvc,
	"hevc":      Hvc,
	"hevC":      Hvc,
	"hvc":       Hvc,
	"mp4_hvc":   Hvc,
	"mp4-hvc":   Hvc,
	"hvc1":      Hvc,
	"dvh1":      Hvc,
	"hvc2":      Hvc,
	"hvc3":      Hvc,
	"v_hvc":     Hvc,
	"v_hvc1":    Hvc,
	"hev":       Hev, // ↓ HEVC video with parameter sets also in the Samples, not supported on macOS
	"mp4_hev":   Hev,
	"mp4-hev":   Hev,
	"hev1":      Hev,
	"hev2":      Hev,
	"hev3":      Hev,
	"dvhe":      Hev,
	"h266":      Vvc, // ↓ Versatile Video Coding (VVC) in MP4 container
	"h.266":     Vvc,
	"vvc":       Vvc,
	"vvc1":      Vvc,
	"vvcC":      Vvc,
	"evc":       Evc, // ↓ MPEG-5 Essential Video Coding (EVC) in MP4 container
	"evc1":      Evc,
	"evcC":      Evc,
	"vp8":       Vp8, // ↓ Google VP8 in WebM container
	"vp08":      Vp8,
	"vp80":      Vp8,
	"vp9":       Vp9, // ↓ Google VP9 in WebM container
	"vp09":      Vp9,
	"vp90":      Vp9,
	"av1":       Av1, // ↓ AV1 (AOMedia Video 1) in MP4 container
	"av01":      Av1,
	"mp4_av1":   Av1,
	"mp4_av01":  Av1,
	"iso_av1":   Av1,
	"iso_av01":  Av1,
	"mp4-av1":   Av1,
	"mp4-av01":  Av1,
	"iso-av1":   Av1,
	"iso-av01":  Av1,
	"webm_av1":  WebmAv1, // ↓ AV1 (AOMedia Video 1) in WebM container
	"webm_av01": WebmAv1,
	"webm-av1":  WebmAv1,
	"webm-av01": WebmAv1,
	"webmav1":   WebmAv1,
	"webmav01":  WebmAv1,
	"mkv_av1":   MkvAv1, // ↓ AV1 (AOMedia Video 1) in Matroska container
	"mkv_av01":  MkvAv1,
	"mkv-av1":   MkvAv1,
	"mkv-av01":  MkvAv1,
	"mkv":       MkvAv1,
	"mk1":       MkvAv1,
	"mkvav1":    MkvAv1,
	"mkvav01":   MkvAv1,
	"mkv1":      MkvAv1,
	"ogg":       Theora, // ↓ Theora video in OGG container
	"ogv":       Theora,
	"mp4":       Mp4, // ↓ Unknown codec in MP4 container
	"mpeg4":     Mp4,
	"webm":      Webm, // ↓ Unknown codec in WebM container
	"WebM":      Webm,
}

// Standards maps names to standardized formats.
type Standards map[string]Type
