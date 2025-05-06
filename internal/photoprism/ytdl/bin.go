package ytdl

import (
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

var (
	Bin        = ""
	FFmpegBin  = ""
	FFprobeBin = ""
)

// FindBin returns the YouTube / M38U video downloader binary name.
func FindBin() string {
	if Bin == "" {
		Bin = config.FindBin("yt-dlp", "yt-dl", "youtube-dl", "dl")
	}

	return Bin
}

// FindFFmpegBin returns the "ffmpeg" command binary name.
func FindFFmpegBin() string {
	if FFmpegBin == "" {
		FFmpegBin = config.FindBin(encode.FFmpegBin)
	}

	return FFmpegBin
}

// FindFFprobeBin returns the "ffprobe" command binary name.
func FindFFprobeBin() string {
	if FFprobeBin == "" {
		FFprobeBin = config.FindBin(encode.FFprobeBin)
	}

	return FFprobeBin
}
