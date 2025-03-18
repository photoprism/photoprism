package v4l

import (
	"os/exec"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

// TranscodeToAvcCmd returns the FFmpeg command for hardware-accelerated transcoding to MPEG-4 AVC.
func TranscodeToAvcCmd(srcName, destName string, opt encode.Options) *exec.Cmd {
	// ffmpeg -hide_banner -h encoder=h264_v4l2m2m
	return exec.Command(
		opt.Bin,
		"-y",
		"-strict", "-2",
		"-i", srcName,
		"-c:v", opt.Encoder.String(),
		"-map", opt.MapVideo,
		"-map", opt.MapAudio,
		"-c:a", "aac",
		"-vf", opt.VideoFilter(encode.FormatYUV420P),
		"-num_output_buffers", "72",
		"-num_capture_buffers", "64",
		"-max_muxing_queue_size", "1024",
		"-crf", "23",
		"-r", "30",
		"-b:v", opt.DestBitrate,
		"-f", "mp4",
		"-movflags", "+faststart", // puts headers at the beginning for faster streaming
		destName,
	)
}
