package vaapi

import (
	"os/exec"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

// TranscodeToAvcCmd returns the FFmpeg command for hardware-accelerated transcoding to MPEG-4 AVC.
func TranscodeToAvcCmd(srcName, destName string, opt encode.Options) *exec.Cmd {
	if opt.Device != "" {
		return exec.Command(
			opt.Bin,
			"-hide_banner",
			"-y",
			"-strict", "-2",
			"-hwaccel", "vaapi",
			"-hwaccel_device", opt.Device,
			"-i", srcName,
			"-c:a", "aac",
			"-vf", opt.VideoFilter(encode.FormatNV12),
			"-c:v", opt.Encoder.String(),
			"-map", opt.MapVideo,
			"-map", opt.MapAudio,
			"-ignore_unknown",
			"-qp", opt.QpQuality(),
			"-f", "mp4",
			"-movflags", opt.MovFlags,
			"-map_metadata", opt.MapMetadata,
			destName,
		)
	} else {
		return exec.Command(
			opt.Bin,
			"-hide_banner",
			"-y",
			"-strict", "-2",
			"-hwaccel", "vaapi",
			"-i", srcName,
			"-c:a", "aac",
			"-vf", opt.VideoFilter(encode.FormatNV12),
			"-c:v", opt.Encoder.String(),
			"-map", opt.MapVideo,
			"-map", opt.MapAudio,
			"-ignore_unknown",
			"-qp", opt.QpQuality(),
			"-f", "mp4",
			"-movflags", opt.MovFlags,
			"-map_metadata", opt.MapMetadata,
			destName,
		)
	}
}
