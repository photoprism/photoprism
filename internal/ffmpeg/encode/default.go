package encode

import (
	"os"
	"strings"

	"github.com/photoprism/photoprism/pkg/fs"
)

// defaultAvcEncoder is the default FFmpeg AVC encoder if it has already been determined.
var defaultAvcEncoder = Encoder("")

// DefaultAvcEncoder determines and returns the default FFmpeg AVC encoder type:
func DefaultAvcEncoder() Encoder {
	if defaultAvcEncoder != "" {
		return defaultAvcEncoder
	}

	switch {
	// Default to Nvidia AVC encoder if the NVIDIA_DRIVER_CAPABILITIES variable is set and contains "video":
	case fs.DeviceExists("/dev/nvidia0") &&
		strings.Contains(os.Getenv("NVIDIA_DRIVER_CAPABILITIES"), "video") &&
		!strings.Contains(os.Getenv("PHOTOPRISM_INIT"), "ffmpeg"):
		defaultAvcEncoder = NvidiaAvc
	// Otherwise, use the standard software AVC encoder:
	default:
		defaultAvcEncoder = SoftwareAvc
	}

	return defaultAvcEncoder
}
