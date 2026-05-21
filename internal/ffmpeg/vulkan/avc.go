package vulkan

import (
	"os/exec"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

// TranscodeToAvcCmd returns the FFmpeg command for hardware-accelerated transcoding to MPEG-4 AVC via the Vulkan video extensions.
func TranscodeToAvcCmd(srcName, destName string, opt encode.Options) *exec.Cmd {
	// Vulkan needs a named hardware device that filters and the encoder can attach to;
	// "vk" is the local alias, and an optional physical device selector is appended when set.
	hwDevice := "vulkan=vk"
	if opt.Device != "" {
		hwDevice = "vulkan=vk:" + opt.Device
	}

	// Scale + format conversion happens in system memory, then hwupload moves frames into the Vulkan device.
	videoFilter := opt.VideoFilter(encode.FormatNV12) + ",hwupload"

	// #nosec G204 -- command arguments are built from validated options and paths.
	return exec.Command(
		opt.Bin,
		"-hide_banner",
		"-y",
		"-strict", "-2",
		"-init_hw_device", hwDevice,
		"-filter_hw_device", "vk",
		"-i", srcName,
		"-c:a", "aac",
		"-vf", videoFilter,
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
