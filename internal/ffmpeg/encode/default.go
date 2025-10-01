package encode

import (
	"os"
	"strings"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// defaultAvcEncoder is the default FFmpeg AVC encoder if it has already been determined.
var defaultAvcEncoder = Encoder("")

// DefaultAvcEncoder determines and returns the default FFmpeg AVC encoder type:
func DefaultAvcEncoder() Encoder {
	if defaultAvcEncoder != "" {
		return defaultAvcEncoder
	}

	// See: https://docs.photoprism.app/getting-started/config-options/#docker-image
	init := os.Getenv("PHOTOPRISM_INIT")

	// See: https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/docker-specialized.html
	dCap := os.Getenv("NVIDIA_DRIVER_CAPABILITIES")
	vDev := os.Getenv("NVIDIA_VISIBLE_DEVICES")

	// Check if a GPU is shared through the NVIDIA Container Toolkit.
	switch {
	case fs.DeviceExists("/dev/nvidia0") && !strings.Contains(init, "ffmpeg") &&
		(dCap == "video" || dCap == "all") && (txt.IsUInt(vDev) || vDev == "all"):
		// Enable Nvidia AVC encoder.
		defaultAvcEncoder = NvidiaAvc
	default:
		// Use AVC software encoder.
		defaultAvcEncoder = SoftwareAvc
	}

	return defaultAvcEncoder
}
