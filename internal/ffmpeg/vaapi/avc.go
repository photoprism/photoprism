package vaapi

import (
	"os/exec"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

// TranscodeToAvcCmd returns the FFmpeg command for hardware-accelerated transcoding to MPEG-4 AVC.
func TranscodeToAvcCmd(srcName, destName string, opt encode.Options) *exec.Cmd {
	// FFmpeg 8 no longer derives a filter device from "-hwaccel vaapi" alone, so the
	// "format=nv12,hwupload" step fails with "A hardware device reference is required
	// to upload frames to." We therefore initialize a named VAAPI device and reference
	// it for both decoding ("-hwaccel_device") and filtering ("-filter_hw_device").
	// Without a configured device path FFmpeg auto-detects the default render node.
	initDevice := "vaapi=va"
	if opt.Device != "" {
		initDevice = "vaapi=va:" + opt.Device
	}

	// #nosec G204 -- command arguments are built from validated options and paths.
	return exec.Command(
		opt.Bin,
		"-hide_banner",
		"-y",
		"-strict", "-2",
		"-init_hw_device", initDevice,
		"-hwaccel", "vaapi",
		"-hwaccel_device", "va",
		"-filter_hw_device", "va",
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
