package ffmpeg

import (
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

// ExtractImageCmd extracts a still image from the specified source video file.
func ExtractImageCmd(videoName, imageName string, opt *encode.Options) *exec.Cmd {
	imageExt := strings.ToLower(filepath.Ext(imageName))

	switch imageExt {
	case ".png":
		return ExtractPngImageCmd(videoName, imageName, opt)
	default:
		return ExtractJpegImageCmd(videoName, imageName, opt)
	}
}

// ExtractJpegImageCmd extracts a JPEG still image from the specified source video file.
func ExtractJpegImageCmd(videoName, imageName string, opt *encode.Options) *exec.Cmd {
	// TODO: Adjust command flags for correct colors with HDR10-encoded HEVC videos,
	//       see https://github.com/photoprism/photoprism/issues/4488.
	// Unfortunately, this filter would render thumbnails of non-HDR videos too dark:
	// "-vf", "zscale=t=linear:npl=100,format=gbrpf32le,zscale=p=bt709,tonemap=tonemap=gamma:desat=0,zscale=t=bt709:m=bt709:r=tv,format=yuv420p",
	return exec.Command(
		opt.Bin,
		"-hide_banner",
		"-loglevel", "error",
		"-y", "-strict", "-2", // support new video codecs
		"-hwaccel", "none", // disable hardware acceleration
		"-err_detect", "ignore_err", // ignore errors
		"-ss", opt.SeekOffset, // open video at this position
		"-i", videoName, // input video file name
		"-ss", opt.TimeOffset, // extract image at this position
		// "-map", opt.MapVideo, "-an", "-sn", "-dn", // map streams (seems not required)
		// "-skip_frame", "nokey", // skip non-keyframes
		"-vf", "setparams=range=tv:color_primaries=bt709:color_trc=bt709:colorspace=bt709,scale=trunc(iw/2)*2:trunc(ih/2)*2,setsar=1,format=yuvj422p",
		"-frames:v", "1", // extract one frame
		imageName, // output image file name
	)
}

// ExtractPngImageCmd extracts a PNG still image from the specified source video file.
func ExtractPngImageCmd(videoName, imageName string, opt *encode.Options) *exec.Cmd {
	return exec.Command(
		opt.Bin,
		"-hide_banner",
		"-loglevel", "error",
		"-y", "-strict", "-2", // support new video codecs
		"-hwaccel", "none", // disable hardware acceleration
		"-err_detect", "ignore_err", // ignore errors
		"-ss", opt.SeekOffset, // open video at this position
		"-i", videoName, // input video file name
		"-ss", opt.TimeOffset, // extract image at this position
		// "-map", opt.MapVideo, "-an", "-sn", "-dn", // map streams (seems not required)
		// "-skip_frame", "nokey", // skip non-keyframes
		"-vf", "scale=trunc(iw/2)*2:trunc(ih/2)*2,setsar=1",
		"-frames:v", "1", // extract one frame
		imageName, // output image file name
	)
}
