package video

import (
	"time"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
)

// Info represents video file information.
type Info struct {
	FileName      string
	FileSize      int64
	FileType      fs.Type
	MediaType     media.Type
	ThumbOffset   int64
	ThumbMimeType string
	VideoOffset   int64
	VideoMimeType string
	VideoType     Type
	VideoCodec    Codec
	VideoWidth    int
	VideoHeight   int
	Duration      time.Duration
	Frames        int
	FPS           float64
	Tracks        int
	Encrypted     bool
	FastStart     bool
	Compatible    bool
}

// NewInfo returns a new Info struct with default values.
func NewInfo() Info {
	return Info{
		FileType:    fs.TypeUnknown,
		FileSize:    -1,
		MediaType:   media.Unknown,
		ThumbOffset: -1,
		VideoOffset: -1,
		VideoType:   Unknown,
		VideoCodec:  CodecUnknown,
		FPS:         0.0,
	}
}

// VideoSize returns the size of the embedded video, if possible.
func (info Info) VideoSize() int64 {
	if info.FileSize < 0 || info.VideoOffset < 0 {
		return 0
	}
	return info.FileSize - info.VideoOffset
}

// VideoContentType composes the video content type from its mime type and codec.
func (info Info) VideoContentType() string {
	if info.VideoMimeType == "" {
		return ""
	}
	return ContentType(info.VideoMimeType, info.VideoCodec)
}
