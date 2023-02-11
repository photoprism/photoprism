package ffmpeg

import "time"

// PreviewTimeOffset returns an appropriate time offset depending on the duration for extracting a preview image.
func PreviewTimeOffset(d time.Duration) string {
	// Default.
	result := "00:00:00.001"

	// If the video is long enough, don't use the first frames to avoid completely
	// black or white thumbnails in case there is an effect or intro.
	switch {
	case d > time.Hour:
		result = "00:02:30.000"
	case d > 10*time.Minute:
		result = "00:01:00.000"
	case d > 3*time.Minute:
		result = "00:00:30.000"
	case d > time.Minute:
		result = "00:00:09.000"
	case d > time.Millisecond*3100:
		result = "00:00:03.000"
	}

	return result
}
