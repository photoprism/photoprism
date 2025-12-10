package nvidia

import (
	"os/exec"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

// TranscodeToAvcCmd returns the FFmpeg command for hardware-accelerated transcoding to MPEG-4 AVC.
func TranscodeToAvcCmd(srcName, destName string, opt encode.Options) *exec.Cmd {
	// ffmpeg -hide_banner -h encoder=h264_nvenc
	// #nosec G204 -- command arguments are built from validated options and paths.
	return exec.Command(
		opt.Bin,
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
		"-preset", opt.Preset,
		"-pixel_format", "yuv420p",
		"-gpu", "any",
		"-vf", opt.VideoFilter(encode.FormatYUV420P),
		"-rc:v", "constqp",
		"-cq", opt.CqQuality(),
		"-tune", "2",
		"-profile:v", "1",
		"-level:v", "auto",
		"-coder:v", "1",
		"-f", "mp4",
		"-movflags", opt.MovFlags,
		"-map_metadata", opt.MapMetadata,
		destName,
	)
}
