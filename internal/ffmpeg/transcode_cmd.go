package ffmpeg

import (
	"fmt"
	"os/exec"

	"github.com/photoprism/photoprism/internal/ffmpeg/apple"
	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/internal/ffmpeg/intel"
	"github.com/photoprism/photoprism/internal/ffmpeg/nvidia"
	"github.com/photoprism/photoprism/internal/ffmpeg/v4l"
	"github.com/photoprism/photoprism/internal/ffmpeg/vaapi"
	"github.com/photoprism/photoprism/pkg/fs"
)

// TranscodeCmd returns the FFmpeg command for transcoding existing video files to MPEG-4 AVC.
func TranscodeCmd(srcName, destName string, opt encode.Options) (cmd *exec.Cmd, useMutex bool, err error) {
	if srcName == "" {
		return nil, false, fmt.Errorf("empty source filename")
	} else if destName == "" {
		return nil, false, fmt.Errorf("empty destination filename")
	}

	// Prevents multiple videos from being transcoded at the same time.
	useMutex = true

	// Use the default binary name if no name is specified.
	if opt.Bin == "" {
		opt.Bin = DefaultBin
	}

	// Always use software encoder for transcoding animated pictures into videos.
	if fs.TypeAnimated[fs.FileType(srcName)] != "" {
		cmd = exec.Command(
			opt.Bin,
			"-y",
			"-strict", "-2",
			"-i", srcName,
			"-pix_fmt", encode.FormatYUV420P.String(),
			"-vf", "scale='trunc(iw/2)*2:trunc(ih/2)*2'",
			"-f", "mp4",
			"-movflags", "+faststart", // puts headers at the beginning for faster streaming
			destName,
		)

		return cmd, useMutex, nil
	}

	// Log encoder name if it is not the default.
	if opt.Encoder != encode.SoftwareAvc {
		log.Infof("convert: ffmpeg encoder %s selected", opt.Encoder.String())
	}

	// Transcode video with selected encoder.
	switch opt.Encoder {
	case encode.IntelAvc:
		cmd = intel.TranscodeToAvcCmd(srcName, destName, opt)
	case encode.AppleAvc:
		cmd = apple.TranscodeToAvcCmd(srcName, destName, opt)
	case encode.VaapiAvc:
		cmd = vaapi.TranscodeToAvcCmd(srcName, destName, opt)
	case encode.NvidiaAvc:
		cmd = nvidia.TranscodeToAvcCmd(srcName, destName, opt)
	case encode.V4LAvc:
		cmd = v4l.TranscodeToAvcCmd(srcName, destName, opt)
	default:
		cmd = encode.TranscodeToAvcCmd(srcName, destName, opt)
	}

	return cmd, useMutex, nil
}
