package encode

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPreviewTimeOffset(t *testing.T) {
	t.Run("Second", func(t *testing.T) {
		assert.Equal(t, "00:00:00.001", PreviewTimeOffset(time.Second))
	})
	t.Run("Minute", func(t *testing.T) {
		assert.Equal(t, "00:00:03.000", PreviewTimeOffset(time.Minute))
	})
	t.Run("ThreeMinutes", func(t *testing.T) {
		assert.Equal(t, "00:00:09.000", PreviewTimeOffset(3*time.Minute))
	})
	t.Run("FiveMinutes", func(t *testing.T) {
		assert.Equal(t, "00:00:30.000", PreviewTimeOffset(5*time.Minute))
	})
	t.Run("FifteenMinutes", func(t *testing.T) {
		assert.Equal(t, "00:01:00.000", PreviewTimeOffset(15*time.Minute))
	})
	t.Run("HalfHour", func(t *testing.T) {
		assert.Equal(t, "00:01:00.000", PreviewTimeOffset(30*time.Minute))
	})
	t.Run("Hour", func(t *testing.T) {
		assert.Equal(t, "00:01:00.000", PreviewTimeOffset(time.Hour))
	})
	t.Run("ThreeHours", func(t *testing.T) {
		assert.Equal(t, "00:02:30.000", PreviewTimeOffset(3*time.Hour))
	})
}
