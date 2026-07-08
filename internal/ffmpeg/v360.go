package ffmpeg

import (
	"fmt"
	"os/exec"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

// V360DualFisheyeToEquirect returns the FFmpeg "v360" filter that dewarps a single-frame,
// side-by-side dual-fisheye source (e.g. an Insta360 .insp/.insv original) to equirectangular.
// fov is the per-lens field of view in degrees (~204 for typical Insta360 X-series cameras).
func V360DualFisheyeToEquirect(fov int) string {
	return fmt.Sprintf("v360=input=dfisheye:output=e:ih_fov=%d:iv_fov=%d", fov, fov)
}

// DewarpDualFisheyeToJpegCmd returns the command that dewarps a dual-fisheye still image
// (e.g. an Insta360 .insp file, which FFmpeg decodes as JPEG) to an equirectangular JPEG.
func DewarpDualFisheyeToJpegCmd(inputName, jpegName string, fov int, opt *encode.Options) *exec.Cmd {
	// #nosec G204 -- paths and flags are created by the application, not user input.
	return exec.Command(
		opt.Bin,
		"-hide_banner",
		"-loglevel", "error",
		"-y",
		"-i", inputName, // input dual-fisheye image
		"-vf", V360DualFisheyeToEquirect(fov), // dewarp to equirectangular
		"-frames:v", "1", // write a single frame
		jpegName, // output equirectangular JPEG
	)
}
