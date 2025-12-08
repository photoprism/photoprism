package encode

import (
	"fmt"
	"time"

	"github.com/photoprism/photoprism/pkg/fs"
)

// Options represents FFmpeg encoding options.
type Options struct {
	Bin         string        // FFmpeg binary filename, e.g. /usr/bin/ffmpeg
	Container   fs.Type       // Multimedia Container File Format
	Encoder     Encoder       // Supported FFmpeg output Encoder
	SizeLimit   int           // Maximum width and height of the output video file in pixels.
	Quality     int           // See https://ffmpeg.org/ffmpeg-codecs.html
	Preset      string        // See https://trac.ffmpeg.org/wiki/Encode/H.264#Preset
	Device      string        // See https://trac.ffmpeg.org/wiki/Limiting%20the%20output%20bitrate
	MapVideo    string        // See https://trac.ffmpeg.org/wiki/Map#Videostreamsonly
	MapAudio    string        // See https://trac.ffmpeg.org/wiki/Map#Audiostreamsonly
	MapMetadata string        // See https://ffmpeg.org/ffmpeg.html
	SeekOffset  string        // See https://trac.ffmpeg.org/wiki/Seeking and https://ffmpeg.org/ffmpeg-utils.html#time-duration-syntax
	TimeOffset  string        // See https://trac.ffmpeg.org/wiki/Seeking and https://ffmpeg.org/ffmpeg-utils.html#time-duration-syntax
	Duration    time.Duration // See https://ffmpeg.org/ffmpeg.html#Main-options
	MovFlags    string
	Title       string
	Description string
	Comment     string
	Author      string
	Created     time.Time
	Force       bool
}

// NewVideoOptions creates and returns new FFmpeg video transcoding options.
func NewVideoOptions(ffmpegBin string, encoder Encoder, sizeLimit, quality int, preset, device, mapVideo, mapAudio string) Options {
	if ffmpegBin == "" {
		ffmpegBin = FFmpegBin
	}

	if encoder == "" {
		encoder = DefaultAvcEncoder()
	}

	switch {
	case sizeLimit < 1:
		sizeLimit = 1920
	case sizeLimit > 15360:
		sizeLimit = 15360
	}

	switch {
	case quality <= 0:
		quality = DefaultQuality
	case quality < WorstQuality:
		quality = WorstQuality
	case quality >= BestQuality:
		quality = BestQuality
	}

	if preset == "" {
		preset = PresetFast
	}

	if mapVideo == "" {
		mapVideo = DefaultMapVideo
	}

	if mapAudio == "" {
		mapAudio = DefaultMapAudio
	}

	return Options{
		Bin:         ffmpegBin,
		Container:   fs.VideoMp4,
		Encoder:     encoder,
		SizeLimit:   sizeLimit,
		Quality:     quality,
		Preset:      preset,
		Device:      device,
		MapVideo:    mapVideo,
		MapAudio:    mapAudio,
		MapMetadata: DefaultMapMetadata,
		MovFlags:    MovFlags,
	}
}

// NewRemuxOptions creates and returns new video remux options.
func NewRemuxOptions(ffmpegBin string, container fs.Type, force bool) Options {
	if ffmpegBin == "" {
		ffmpegBin = FFmpegBin
	}

	if container == "" {
		container = fs.VideoMp4
	}

	return Options{
		Bin:         ffmpegBin,
		Container:   fs.VideoMp4,
		MapVideo:    DefaultMapVideo,
		MapAudio:    DefaultMapAudio,
		MapMetadata: DefaultMapMetadata,
		MovFlags:    MovFlags,
		Force:       force,
	}
}

// NewPreviewImageOptions generates encoding options for extracting a video preview image.
func NewPreviewImageOptions(ffmpegBin string, videoDuration time.Duration) *Options {
	return &Options{
		Bin:         ffmpegBin,
		MapVideo:    DefaultMapVideo,
		MapAudio:    DefaultMapAudio,
		MapMetadata: DefaultMapMetadata,
		SeekOffset:  PreviewSeekOffset(videoDuration),
		TimeOffset:  PreviewTimeOffset(videoDuration),
	}
}

// VideoFilter returns the FFmpeg video filter string based on the size limit in pixels and the pixel format.
func (o *Options) VideoFilter(format PixelFormat) string {
	// scale specifies the FFmpeg downscale filter, see http://trac.ffmpeg.org/wiki/Scaling.
	switch format {
	case "":
		return fmt.Sprintf("scale='if(gte(iw,ih), min(%d, iw), -2):if(gte(iw,ih), -2, min(%d, ih))'", o.SizeLimit, o.SizeLimit)
	case FormatQSV:
		return fmt.Sprintf("scale_qsv=w='if(gte(iw,ih), min(%d, iw), -1)':h='if(gte(iw,ih), -1, min(%d, ih))':format=nv12", o.SizeLimit, o.SizeLimit)
	}

	return fmt.Sprintf("scale='if(gte(iw,ih), min(%d, iw), -2):if(gte(iw,ih), -2, min(%d, ih))',format=%s", o.SizeLimit, o.SizeLimit, format)
}

// QvQuality  returns the video encoding quality as "-q:v" parameter string.
func (o *Options) QvQuality() string {
	return QvQuality(o.Quality)
}

// GlobalQuality returns the video encoding quality as "-global_quality" parameter string.
func (o *Options) GlobalQuality() string {
	return GlobalQuality(o.Quality)
}

// CrfQuality returns the video encoding quality as "-crf" parameter string.
func (o *Options) CrfQuality() string {
	return CrfQuality(o.Quality)
}

// QpQuality returns the video encoding quality as "-qp" parameter string.
func (o *Options) QpQuality() string {
	return QpQuality(o.Quality)
}

// CqQuality returns the video encoding quality as "-cq" parameter string.
func (o *Options) CqQuality() string {
	return CqQuality(o.Quality)
}
