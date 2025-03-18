package encode

import (
	"fmt"
	"time"
)

// Options represents FFmpeg encoding options.
type Options struct {
	Bin         string        // FFmpeg binary filename, e.g. /usr/bin/ffmpeg.
	Encoder     Encoder       // Supported FFmpeg output Encoder.
	DestSize    int           // Maximum width and height of the output video file in pixels.
	DestBitrate string        // See https://trac.ffmpeg.org/wiki/Limiting%20the%20output%20bitrate.
	MapVideo    string        // See https://trac.ffmpeg.org/wiki/Map#Videostreamsonly.
	MapAudio    string        // See https://trac.ffmpeg.org/wiki/Map#Audiostreamsonly.
	TimeOffset  string        // See https://trac.ffmpeg.org/wiki/Seeking and https://ffmpeg.org/ffmpeg-utils.html#time-duration-syntax.
	Duration    time.Duration // See https://ffmpeg.org/ffmpeg.html#Main-options.
}

// NewPreviewImageOptions generates encoding options for extracting a video preview image.
func NewPreviewImageOptions(ffmpegBin string, videoDuration time.Duration) *Options {
	return &Options{
		Bin:        ffmpegBin,
		TimeOffset: PreviewTimeOffset(videoDuration),
	}
}

// VideoFilter returns the FFmpeg video filter string based on the size limit in pixels and the pixel format.
func (o *Options) VideoFilter(format PixelFormat) string {
	// scale specifies the FFmpeg downscale filter, see http://trac.ffmpeg.org/wiki/Scaling.
	if format == "" {
		return fmt.Sprintf("scale='if(gte(iw,ih), min(%d, iw), -2):if(gte(iw,ih), -2, min(%d, ih))'", o.DestSize, o.DestSize)
	} else if format == FormatQSV {
		return fmt.Sprintf("scale_qsv=w='if(gte(iw,ih), min(%d, iw), -1)':h='if(gte(iw,ih), -1, min(%d, ih))':format=nv12", o.DestSize, o.DestSize)
	} else {
		return fmt.Sprintf("scale='if(gte(iw,ih), min(%d, iw), -2):if(gte(iw,ih), -2, min(%d, ih))',format=%s", o.DestSize, o.DestSize, format)
	}
}
