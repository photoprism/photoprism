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
	return exec.Command(opt.Bin, "-y", "-strict", "-2", "-ss", opt.TimeOffset, "-i", videoName, "-vframes", "1", imageName)
}

// ExtractPngImageCmd extracts a PNG still image from the specified source video file.
func ExtractPngImageCmd(videoName, imageName string, opt *encode.Options) *exec.Cmd {
	return exec.Command(opt.Bin, "-y", "-strict", "-2", "-ss", opt.TimeOffset, "-i", videoName, "-vframes", "1", imageName)
}
