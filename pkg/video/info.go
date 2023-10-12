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

// VideoBitrate returns the bitrate of the embedded video in MBit/s.
func (info Info) VideoBitrate() float64 {
	videoSize := info.VideoSize()

	// Return 0 if video size or duration are unknown.
	if videoSize <= 0 || info.Duration <= 0 {
		return 0
	}

	// Divide number of bits through the duration in seconds.
	return ((float64(videoSize) * 8) / info.Duration.Seconds()) / 1e6
}

// VideoContentType composes the video content type from its mime type and codec.
func (info Info) VideoContentType() string {
	if info.VideoMimeType == "" {
		return ContentTypeDefault
	}

	return ContentType(info.VideoMimeType, info.VideoCodec)
}

// VideoFileExt returns the appropriate video file extension based on the mime type and defaults to fs.ExtMP4 otherwise.
func (info Info) VideoFileExt() string {
	switch info.VideoMimeType {
	case fs.MimeTypeMOV:
		return fs.ExtMOV
	default:
		return fs.ExtMP4
	}
}

// VideoFileType returns the video type based on the mime type and defaults to fs.VideoMP4 otherwise.
func (info Info) VideoFileType() fs.Type {
	switch info.VideoMimeType {
	case fs.MimeTypeMOV:
		return fs.VideoMOV
	default:
		return fs.VideoMP4
	}
}
