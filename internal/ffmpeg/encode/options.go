package encode

import (
	"fmt"
	"time"
)

// Options represents FFmpeg encoding options.
type Options struct {
	Bin          string        // FFmpeg binary filename, e.g. /usr/bin/ffmpeg.
	Encoder      Encoder       // Supported FFmpeg output Encoder.
	SizeLimit    int           // Maximum width and height of the output video file in pixels.
	BitrateLimit string        // See https://trac.ffmpeg.org/wiki/Limiting%20the%20output%20bitrate.
	MapVideo     string        // See https://trac.ffmpeg.org/wiki/Map#Videostreamsonly.
	MapAudio     string        // See https://trac.ffmpeg.org/wiki/Map#Audiostreamsonly.
	TimeOffset   string        // See https://trac.ffmpeg.org/wiki/Seeking and https://ffmpeg.org/ffmpeg-utils.html#time-duration-syntax.
	Duration     time.Duration // See https://ffmpeg.org/ffmpeg.html#Main-options.
	MovFlags     string
}

// NewVideoOptions creates and returns new FFmpeg video transcoding options.
func NewVideoOptions(ffmpegBin string, encoder Encoder, sizeLimit int, bitrateLimit, mapVideo, mapAudio string) Options {
	if ffmpegBin == "" {
		ffmpegBin = FFmpegBin
	}

	if encoder == "" {
		encoder = DefaultAvcEncoder()
	}

	if sizeLimit < 1 {
		sizeLimit = 1920
	} else if sizeLimit > 15360 {
		sizeLimit = 15360
	}

	if bitrateLimit == "" {
		bitrateLimit = "50M"
	}

	if mapVideo == "" {
		mapVideo = MapVideo
	}

	if mapAudio == "" {
		mapAudio = MapAudio
	}

	return Options{
		Bin:          ffmpegBin,
		Encoder:      encoder,
		SizeLimit:    sizeLimit,
		BitrateLimit: bitrateLimit,
		MapVideo:     mapVideo,
		MapAudio:     mapAudio,
		MovFlags:     MovFlags,
	}
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
		return fmt.Sprintf("scale='if(gte(iw,ih), min(%d, iw), -2):if(gte(iw,ih), -2, min(%d, ih))'", o.SizeLimit, o.SizeLimit)
	} else if format == FormatQSV {
		return fmt.Sprintf("scale_qsv=w='if(gte(iw,ih), min(%d, iw), -1)':h='if(gte(iw,ih), -1, min(%d, ih))':format=nv12", o.SizeLimit, o.SizeLimit)
	} else {
		return fmt.Sprintf("scale='if(gte(iw,ih), min(%d, iw), -2):if(gte(iw,ih), -2, min(%d, ih))',format=%s", o.SizeLimit, o.SizeLimit, format)
	}
}
