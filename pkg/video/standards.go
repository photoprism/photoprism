package video

// Types maps identifiers to standards.
var Types = Standards{
	"":      AVC,
	"mp4":   MP4,
	"mpeg4": MP4,
	"avc":   AVC,
	"avc1":  AVC,
	"hvc":   HEVC,
	"hvc1":  HEVC,
	"hevc":  HEVC,
	"vvc":   VVC,
	"vvc1":  VVC,
	"av1":   AV1,
	"av01":  AV1,
}

// Standards maps names to standardized formats.
type Standards map[string]Type
