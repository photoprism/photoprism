package encode

import "time"

// PreviewTimeOffset returns a time offset depending on the video duration for extracting a cover image,
// see https://trac.ffmpeg.org/wiki/Seeking and https://ffmpeg.org/ffmpeg-utils.html#time-duration-syntax.
func PreviewTimeOffset(d time.Duration) string {
	// Default time offset.
	result := "00:00:00.001"

	if d <= 0 {
		return result
	}

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
