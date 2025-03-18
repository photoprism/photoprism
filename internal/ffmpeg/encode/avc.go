package encode

import "os/exec"

// TranscodeToAvcCmd returns the default FFmpeg command for transcoding video files to MPEG-4 AVC.
func TranscodeToAvcCmd(srcName, destName string, opt Options) *exec.Cmd {
	return exec.Command(
		opt.Bin,
		"-y",
		"-strict", "-2",
		"-i", srcName,
		"-c:v", opt.Encoder.String(),
		"-map", opt.MapVideo,
		"-map", opt.MapAudio,
		"-c:a", "aac",
		"-vf", opt.VideoFilter(FormatYUV420P),
		"-max_muxing_queue_size", "1024",
		"-crf", "23",
		"-r", "30",
		"-b:v", opt.DestBitrate,
		"-f", "mp4",
		"-movflags", "+faststart", // puts headers at the beginning for faster streaming
		destName,
	)
}
