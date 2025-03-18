package intel

import (
	"os/exec"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

// TranscodeToAvcCmd returns the FFmpeg command for hardware-accelerated transcoding to MPEG-4 AVC.
func TranscodeToAvcCmd(srcName, destName string, opt encode.Options) *exec.Cmd {
	// ffmpeg -hide_banner -h encoder=h264_qsv
	return exec.Command(
		opt.Bin,
		"-y",
		"-strict", "-2",
		"-hwaccel", "qsv",
		"-hwaccel_output_format", "qsv",
		"-qsv_device", "/dev/dri/renderD128",
		"-i", srcName,
		"-c:a", "aac",
		"-vf", opt.VideoFilter(encode.FormatQSV),
		"-c:v", opt.Encoder.String(),
		"-map", opt.MapVideo,
		"-map", opt.MapAudio,
		"-r", "30",
		"-b:v", opt.DestBitrate,
		"-bitrate", opt.DestBitrate,
		"-f", "mp4",
		"-movflags", "+faststart", // puts headers at the beginning for faster streaming
		destName,
	)
}
