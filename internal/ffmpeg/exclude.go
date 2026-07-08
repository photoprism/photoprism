package ffmpeg

import (
	"sync/atomic"

	"github.com/photoprism/photoprism/pkg/media/video"
)

// excludePtr atomically holds the active FFmpeg exclude list so updates
// from Config.Propagate do not tear concurrent reads on workers and CLI
// commands.
var excludePtr atomic.Pointer[video.Formats]

// DefaultExclude is the comma-separated list of container and codec formats
// that should not be processed by FFmpeg by default.
var DefaultExclude string

func init() {
	f := video.NewFormats(video.CodecMagicYUV, video.CodecVFW)
	excludePtr.Store(&f)
	DefaultExclude = f.String()
}

// Exclude returns the current FFmpeg exclude list. The returned map must not
// be mutated; publish a new list via SetExclude instead.
func Exclude() video.Formats {
	return *excludePtr.Load()
}

// SetExclude replaces the active FFmpeg exclude list atomically. The caller
// must not mutate f after the call.
func SetExclude(f video.Formats) {
	excludePtr.Store(&f)
}
