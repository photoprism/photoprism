package vaapi

import (
	"os/exec"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

// TranscodeToAvcCmd returns the FFmpeg command for hardware-accelerated transcoding to MPEG-4 AVC.
func TranscodeToAvcCmd(srcName, destName string, opt encode.Options) *exec.Cmd {
	return exec.Command(
		opt.Bin,
		"-y",
		"-strict", "-2",
		"-hwaccel", "vaapi",
		"-i", srcName,
		"-c:a", "aac",
		"-vf", opt.VideoFilter(encode.FormatNV12),
		"-c:v", opt.Encoder.String(),
		"-map", opt.MapVideo,
		"-map", opt.MapAudio,
		"-r", "30",
		"-b:v", opt.BitrateLimit,
		"-f", "mp4",
		"-movflags", opt.MovFlags,
		destName,
	)
}
