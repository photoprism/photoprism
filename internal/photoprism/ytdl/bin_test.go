package ytdl

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindBin(t *testing.T) {
	assert.True(t, strings.Contains(FindBin(), "yt-dlp"), "binary filepath should contain 'yt-dlp'")
}

func TestFindFFmpegBin(t *testing.T) {
	assert.True(t, strings.Contains(FindFFmpegBin(), "ffmpeg"), "binary filepath should contain 'ffmpeg'")
}

func TestFindFFprobeBin(t *testing.T) {
	assert.True(t, strings.Contains(FindFFprobeBin(), "ffprobe"), "binary filepath should contain 'ffprobe'")
}
