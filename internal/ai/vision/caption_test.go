package vision

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/media"
)

func TestCaption(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	} else if _, err := net.DialTimeout("tcp", "photoprism-vision:5000", 10*time.Second); err != nil {
		t.Skip("skipping test because photoprism-vision is not running.")
	}

	t.Run("Success", func(t *testing.T) {
		expectedText := "An image of sound waves"

		result, err := Caption("https://dl.photoprism.app/img/artwork/colorwaves-400.jpg", media.SrcRemote)

		assert.NoError(t, err)
		assert.IsType(t, CaptionResult{}, result)
		assert.LessOrEqual(t, float32(0.0), result.Confidence)

		t.Logf("caption: %#v", result)

		assert.Equal(t, expectedText, result.Text)
	})
	t.Run("Invalid", func(t *testing.T) {
		result, err := Caption("", media.SrcLocal)

		assert.Error(t, err)
		assert.IsType(t, CaptionResult{}, result)
		assert.Equal(t, "", result.Text)
		assert.Equal(t, float32(0.0), result.Confidence)
	})
}
