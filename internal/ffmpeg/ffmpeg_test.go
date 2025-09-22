package ffmpeg

import "os"

// Ensure that vendor-specific HW-accelerated test runs are disabled unless explicitly set.
func init() {
	_ = os.Unsetenv("PHOTOPRISM_FFMPEG_ENCODER")
}
