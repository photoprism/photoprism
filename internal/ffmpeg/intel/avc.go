package intel

import (
	"os/exec"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

// TranscodeToAvcCmd returns the FFmpeg command for hardware-accelerated transcoding to MPEG-4 AVC.
func TranscodeToAvcCmd(srcName, destName string, opt encode.Options) *exec.Cmd {
	// ffmpeg -hide_banner -h encoder=h264_qsv
	if opt.Device != "" {
		// #nosec G204 -- command arguments are built from validated options and paths.
		return exec.Command(
			opt.Bin,
			"-hide_banner",
			"-y",
			"-strict", "-2",
			"-hwaccel", "qsv",
			"-hwaccel_device", opt.Device,
			"-hwaccel_output_format", "qsv",
			"-i", srcName,
			"-c:a", "aac",
			"-vf", opt.VideoFilter(encode.FormatQSV),
			"-c:v", opt.Encoder.String(),
			"-map", opt.MapVideo,
			"-map", opt.MapAudio,
			"-ignore_unknown",
			"-preset", opt.Preset,
			"-global_quality", opt.GlobalQuality(),
			"-f", "mp4",
			"-movflags", opt.MovFlags,
			"-map_metadata", opt.MapMetadata,
			destName,
		)
	} else {
		// #nosec G204 -- command arguments are built from validated options and paths.
		return exec.Command(
			opt.Bin,
			"-hide_banner",
			"-y",
			"-strict", "-2",
			"-hwaccel", "qsv",
			"-hwaccel_output_format", "qsv",
			"-i", srcName,
			"-c:a", "aac",
			"-vf", opt.VideoFilter(encode.FormatQSV),
			"-c:v", opt.Encoder.String(),
			"-map", opt.MapVideo,
			"-map", opt.MapAudio,
			"-ignore_unknown",
			"-preset", opt.Preset,
			"-global_quality", opt.GlobalQuality(),
			"-f", "mp4",
			"-movflags", opt.MovFlags,
			"-map_metadata", opt.MapMetadata,
			destName,
		)
	}
}
