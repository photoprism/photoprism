package encode

import "time"

// PreviewSeekOffset returns a seek offset depending on the video duration for extracting a preview image,
// see https://trac.ffmpeg.org/wiki/Seeking and https://ffmpeg.org/ffmpeg-utils.html#time-duration-syntax.
func PreviewSeekOffset(d time.Duration) string {
	// Default time offset.
	result := "00:00:00.000"

	if d <= 0 {
		return result
	}

	// If the video is long enough, don't use the first frames to avoid completely
	// black or white thumbnails in case there is an effect or intro.
	switch {
	case d > time.Hour:
		result = "00:02:28.000"
	case d > 10*time.Minute:
		result = "00:00:58.000"
	case d > 3*time.Minute:
		result = "00:00:28.000"
	}

	return result
}

// PreviewTimeOffset returns a time offset depending on the video duration for extracting a preview image,
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
	}

	return result
}
