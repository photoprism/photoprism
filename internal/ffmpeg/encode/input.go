package encode

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/fs"
)

// InputFormats lists the demuxers FFmpeg may select for a source file whose format is not pinned,
// i.e. the containers, raw bitstreams and animated image formats of the supported file types.
var InputFormats = []string{
	"mov", "matroska", "avi", "mpegts", "mpeg", "mpegvideo", "asf", "flv", "ogg", "mxf", "dv",
	"h264", "hevc", "vvc", "evc", "obu", "ivf", "m4v", "mjpeg", "mpjpeg", "gif", "apng", "webp_pipe",
}

// InputFormatWhitelist returns InputFormats as the value of the FFmpeg "-format_whitelist" option.
func InputFormatWhitelist() string {
	return strings.Join(InputFormats, ",")
}

// InputArgs returns the FFmpeg arguments that open the specified source file.
func InputArgs(fileName string) []string {
	return append(InputFormatArgs(fileName), "-i", fileName)
}

// InputFormatArgs returns the FFmpeg options that select the demuxer for the specified source file. JPEG
// and Insta360 images are read as JPEG and Insta360 videos as MOV/MP4; other files are limited to
// InputFormats, and Motion JPEG files may also be read as a JPEG stream.
func InputFormatArgs(fileName string) []string {
	switch fs.FileType(fileName) {
	case fs.ImageJpeg, fs.ImageInsp:
		return []string{"-f", "jpeg_pipe"}
	case fs.VideoInsv, fs.VideoLrv:
		return []string{"-f", "mov"}
	case fs.VideoMjpeg:
		return []string{"-format_whitelist", InputFormatWhitelist() + ",jpeg_pipe"}
	default:
		return []string{"-format_whitelist", InputFormatWhitelist()}
	}
}
