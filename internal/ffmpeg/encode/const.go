package encode

// FFmpegBin defines the default ffmpeg binary name.
const (
	FFmpegBin  = "ffmpeg"
	FFprobeBin = "ffprobe"
)

// Bitrate limit min, max, and default settings in MBps.
const (
	NoBitrateLimit      = -1
	MinBitrateLimit     = 1
	DefaultBitrateLimit = 60
	MaxBitrateLimit     = 960
)

// Default video and audio track mapping.
const (
	DefaultMapVideo    = "0:v:0"
	DefaultMapAudio    = "0:a:0?"
	DefaultMapMetadata = "0"
)
