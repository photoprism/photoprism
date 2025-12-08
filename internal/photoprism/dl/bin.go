package dl

import (
	"os/exec"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

// Cached binary discovery results.
var (
	YtDlpBin   = ""
	FFmpegBin  = ""
	FFprobeBin = ""
)

// FindYtDlpBin returns the YouTube / M38U video downloader binary name.
func FindYtDlpBin() string {
	if YtDlpBin == "" {
		YtDlpBin, _ = exec.LookPath("yt-dlp")
	}

	return YtDlpBin
}

// FindFFmpegBin returns the "ffmpeg" command binary name.
func FindFFmpegBin() string {
	if FFmpegBin == "" {
		FFmpegBin, _ = exec.LookPath(encode.FFmpegBin)
	}

	return FFmpegBin
}

// FindFFprobeBin returns the "ffprobe" command binary name.
func FindFFprobeBin() string {
	if FFprobeBin == "" {
		FFprobeBin, _ = exec.LookPath(encode.FFprobeBin)
	}

	return FFprobeBin
}
