package encode

import "os/exec"

// TranscodeToAvcCmd returns the default FFmpeg command for transcoding video files to MPEG-4 AVC.
func TranscodeToAvcCmd(srcName, destName string, opt Options) *exec.Cmd {
	return exec.Command(
		opt.Bin,
		"-hide_banner",
		"-y",
		"-strict", "-2",
		"-i", srcName,
		"-c:v", opt.Encoder.String(),
		"-map", opt.MapVideo,
		"-map", opt.MapAudio,
		"-ignore_unknown",
		"-c:a", "aac",
		"-preset", opt.Preset,
		"-vf", opt.VideoFilter(FormatYUV420P),
		"-max_muxing_queue_size", "1024",
		"-crf", opt.CrfQuality(),
		"-f", "mp4",
		"-movflags", opt.MovFlags,
		"-map_metadata", opt.MapMetadata,
		destName,
	)
}
