package ffmpeg

import "os"

// Prevents a PHOTOPRISM_FFMPEG_ENCODER value from the development environment from
// accidentally triggering vendor-specific hardware transcoding tests; real hardware
// runs are opted in explicitly via PHOTOPRISM_FFMPEG_TEST_ENCODER instead.
func init() {
	_ = os.Unsetenv("PHOTOPRISM_FFMPEG_ENCODER")
}
