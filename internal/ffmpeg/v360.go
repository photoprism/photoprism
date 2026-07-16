package ffmpeg

import (
	"fmt"
	"os/exec"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

// V360DualFisheyeToEquirect returns the FFmpeg "v360" filter that dewarps a single-frame,
// side-by-side dual-fisheye source (e.g. an Insta360 .insp/.insv original) to equirectangular.
// fov is the per-lens field of view in degrees and roll is a verified spherical correction.
func V360DualFisheyeToEquirect(fov, roll int) string {
	filter := fmt.Sprintf("v360=input=dfisheye:output=e:ih_fov=%d:iv_fov=%d", fov, fov)

	if roll != 0 {
		filter += fmt.Sprintf(":roll=%d", roll)
	}

	return filter
}

// V360FisheyeToEquirect returns the FFmpeg filter for a single circular fisheye source.
func V360FisheyeToEquirect(fov, roll int) string {
	filter := fmt.Sprintf("v360=input=fisheye:output=e:ih_fov=%d:iv_fov=%d", fov, fov)

	if roll != 0 {
		filter += fmt.Sprintf(":roll=%d", roll)
	}

	return filter
}

// DewarpDualFisheyeToJpegCmd returns the command that dewarps a dual-fisheye still image
// (e.g. an Insta360 .insp file, which FFmpeg decodes as JPEG) to an equirectangular JPEG.
func DewarpDualFisheyeToJpegCmd(inputName, jpegName string, fov, roll int, opt *encode.Options) *exec.Cmd {
	return DewarpFisheyeToJpegCmd(inputName, jpegName, V360DualFisheyeToEquirect(fov, roll), opt)
}

// DewarpFisheyeToJpegCmd dewarps an image with the specified v360 filter to an equirectangular JPEG.
func DewarpFisheyeToJpegCmd(inputName, jpegName, filter string, opt *encode.Options) *exec.Cmd {
	if opt.SizeLimit > 0 {
		scaled := *opt
		scaled.V360 = filter
		filter = scaled.VideoFilter("")
	}

	// #nosec G204 -- paths and flags are created by the application, not user input.
	return exec.Command(
		opt.Bin,
		"-hide_banner",
		"-loglevel", "error",
		"-y",
		"-i", inputName, // input dual-fisheye image
		"-vf", filter, // dewarp to equirectangular
		"-frames:v", "1", // write a single frame
		jpegName, // output equirectangular JPEG
	)
}

// DewarpDualFisheyePairToJpegCmd combines separate left and right lens frames and dewarps them to JPEG.
func DewarpDualFisheyePairToJpegCmd(leftName, rightName, jpegName string, fov, roll int, opt *encode.Options) *exec.Cmd {
	v360Filter := V360DualFisheyeToEquirect(fov, roll)
	if opt.SizeLimit > 0 {
		scaled := *opt
		scaled.V360 = v360Filter
		v360Filter = scaled.VideoFilter("")
	}
	filter := fmt.Sprintf("[0:v:0][1:v:0]hstack=inputs=2:shortest=1,%s[v]", v360Filter)

	// #nosec G204 -- paths and flags are created by the application, not user input.
	return exec.Command(
		opt.Bin,
		"-hide_banner",
		"-loglevel", "error",
		"-y",
		"-i", leftName,
		"-i", rightName,
		"-filter_complex", filter,
		"-map", "[v]",
		"-frames:v", "1",
		jpegName,
	)
}

// DewarpStackedDualFisheyeToJpegCmd rearranges vertically stacked lens frames before dewarping.
func DewarpStackedDualFisheyeToJpegCmd(inputName, jpegName string, fov, roll int, opt *encode.Options) *exec.Cmd {
	v360Filter := V360DualFisheyeToEquirect(fov, roll)
	if opt.SizeLimit > 0 {
		scaled := *opt
		scaled.V360 = v360Filter
		v360Filter = scaled.VideoFilter("")
	}
	filter := fmt.Sprintf("[0:v:0]crop=iw:ih/2:0:0[top];[0:v:0]crop=iw:ih/2:0:ih/2[bottom];[top][bottom]hstack=inputs=2:shortest=1,%s[v]", v360Filter)

	// #nosec G204 -- paths and flags are created by the application, not user input.
	return exec.Command(
		opt.Bin,
		"-hide_banner",
		"-loglevel", "error",
		"-y",
		"-i", inputName,
		"-filter_complex", filter,
		"-map", "[v]",
		"-frames:v", "1",
		jpegName,
	)
}

// DewarpDualFisheyePairToAvcCmd combines separate lens videos and encodes an equirectangular AVC derivative.
func DewarpDualFisheyePairToAvcCmd(leftName, rightName, avcName string, opt encode.Options) *exec.Cmd {
	filter := fmt.Sprintf("[0:v:0][1:v:0]hstack=inputs=2:shortest=1,%s[v]", opt.VideoFilter(encode.FormatYUV420P))

	// #nosec G204 -- paths and flags are created by the application, not user input.
	return exec.Command(
		opt.Bin,
		"-hide_banner",
		"-y",
		"-strict", "-2",
		"-i", leftName,
		"-i", rightName,
		"-filter_complex", filter,
		"-map", "[v]",
		"-map", opt.MapAudio,
		"-ignore_unknown",
		"-c:v", opt.Encoder.String(),
		"-c:a", "aac",
		"-preset", opt.Preset,
		"-max_muxing_queue_size", "1024",
		"-crf", opt.CrfQuality(),
		"-f", "mp4",
		"-movflags", opt.MovFlags,
		"-map_metadata", opt.MapMetadata,
		"-shortest",
		avcName,
	)
}
