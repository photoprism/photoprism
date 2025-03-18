package nvidia

import (
	"os/exec"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

// TranscodeToAvcCmd returns the FFmpeg command for hardware-accelerated transcoding to MPEG-4 AVC.
func TranscodeToAvcCmd(srcName, destName string, opt encode.Options) *exec.Cmd {
	// ffmpeg -hide_banner -h encoder=h264_nvenc
	return exec.Command(
		opt.Bin,
		"-y",
		"-strict", "-2",
		"-hwaccel", "auto",
		"-i", srcName,
		"-pix_fmt", encode.FormatYUV420P.String(),
		"-c:v", opt.Encoder.String(),
		"-map", opt.MapVideo,
		"-map", opt.MapAudio,
		"-c:a", "aac",
		"-preset", "15",
		"-pixel_format", "yuv420p",
		"-gpu", "any",
		"-vf", opt.VideoFilter(encode.FormatYUV420P),
		"-rc:v", "constqp",
		"-cq", "0",
		"-tune", "2",
		"-r", "30",
		"-b:v", opt.DestBitrate,
		"-profile:v", "1",
		"-level:v", "auto",
		"-coder:v", "1",
		"-f", "mp4",
		"-movflags", "+faststart", // puts headers at the beginning for faster streaming
		destName,
	)
}
