package photoprism

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
)

// TestConvertTempPreview verifies direct images are reused and invalid media fails cleanly.
func TestConvertTempPreview(t *testing.T) {
	converter := NewConvert(config.TestConfig())

	t.Run("DirectImage", func(t *testing.T) {
		fileName := filepath.Join("testdata", "flash.jpg")
		mediaFile, err := NewMediaFile(fileName)
		require.NoError(t, err)
		preview, cleanup, err := converter.TempPreview(mediaFile)
		require.NoError(t, err)
		assert.Equal(t, mediaFile.FileName(), preview)
		require.NotNil(t, cleanup)
		cleanup()
	})
	t.Run("Nil", func(t *testing.T) {
		_, cleanup, err := converter.TempPreview(nil)
		require.Error(t, err)
		assert.Nil(t, cleanup)
	})
}
