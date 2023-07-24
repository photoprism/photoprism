package ffmpeg

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPreviewTimeOffset(t *testing.T) {
	assert.Equal(t, "00:00:00.001", PreviewTimeOffset(time.Second))
	assert.Equal(t, "00:00:03.000", PreviewTimeOffset(time.Minute))
	assert.Equal(t, "00:00:09.000", PreviewTimeOffset(3*time.Minute))
	assert.Equal(t, "00:00:30.000", PreviewTimeOffset(5*time.Minute))
	assert.Equal(t, "00:01:00.000", PreviewTimeOffset(15*time.Minute))
	assert.Equal(t, "00:01:00.000", PreviewTimeOffset(30*time.Minute))
	assert.Equal(t, "00:01:00.000", PreviewTimeOffset(time.Hour))
	assert.Equal(t, "00:02:30.000", PreviewTimeOffset(3*time.Hour))
}
