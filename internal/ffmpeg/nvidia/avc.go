package nvidia

import (
	"os/exec"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

// TranscodeToAvcCmd returns the FFmpeg command for hardware-accelerated transcoding to MPEG-4 AVC.
func TranscodeToAvcCmd(srcName, destName string, opt encode.Options) *exec.Cmd {
	// ffmpeg -hide_banner -h encoder=h264_nvenc
	args := []string{
		"-hide_banner",
		"-y",
		"-strict", "-2",
		"-hwaccel", "auto",
		"-i", srcName,
		"-pix_fmt", encode.FormatYUV420P.String(),
		"-c:v", opt.Encoder.String(),
		"-map", opt.MapVideo,
		"-map", opt.MapAudio,
		"-ignore_unknown",
		"-c:a", "aac",
		"-preset", Preset(opt.Preset),
		"-pixel_format", "yuv420p",
		"-gpu", "any",
		"-vf", opt.VideoFilter(encode.FormatYUV420P),
		"-rc:v", "vbr",
		"-cq", opt.CqQuality(),
		"-b:v", "0",
	}

	// The peak bitrate is a rate control target applied per frame at the nominal frame rate, not a hard cap.
	if maxRate := opt.MaxRate(); maxRate != "" {
		args = append(args, "-maxrate", maxRate)
	}

	args = append(args,
		"-tune", "hq",
		"-profile:v", "high",
		"-level:v", "auto",
		"-coder:v", "1",
		"-f", "mp4",
		"-movflags", opt.MovFlags,
		"-map_metadata", opt.MapMetadata,
		destName,
	)

	// #nosec G204 -- command arguments are built from validated options and paths.
	return exec.Command(opt.Bin, args...)
}
