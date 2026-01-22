package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestVideoTrimFastStart verifies which extensions get the faststart flag.
func TestVideoTrimFastStart(t *testing.T) {
	assert.True(t, videoTrimFastStart("clip.mp4"))
	assert.True(t, videoTrimFastStart("clip.MOV"))
	assert.True(t, videoTrimFastStart("clip.m4v"))
	assert.True(t, videoTrimFastStart("clip.qt"))
	assert.False(t, videoTrimFastStart("clip.mkv"))
	assert.False(t, videoTrimFastStart(""))
}
