package ffmpeg

import (
	"os"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

// Prevents a PHOTOPRISM_FFMPEG_ENCODER value from the development environment from
// accidentally triggering vendor-specific hardware transcoding tests; real hardware
// runs are opted in explicitly via PHOTOPRISM_FFMPEG_TEST_ENCODER instead.
func init() {
	_ = os.Unsetenv("PHOTOPRISM_FFMPEG_ENCODER")
}

// testInputWhitelist is the input format option the command builders add before "-i".
var testInputWhitelist = "-format_whitelist " + encode.InputFormatWhitelist()
