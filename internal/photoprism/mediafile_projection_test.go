package photoprism

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/media/projection"
)

// TestMediaFile_VisualProjection verifies explicit, metadata, and generated-sidecar projection paths.
func TestMediaFile_VisualProjection(t *testing.T) {
	t.Run("Explicit", func(t *testing.T) {
		file := &MediaFile{}
		file.SetVisualProjection(projection.Equirectangular)
		assert.Equal(t, projection.Equirectangular, file.VisualProjection(""))
	})
	t.Run("Metadata", func(t *testing.T) {
		file := &MediaFile{}
		assert.Equal(t, projection.Cubestrip, file.VisualProjection(projection.Cubestrip.String()))
	})
	t.Run("GeneratedInspSidecar", func(t *testing.T) {
		conf := config.TestConfig()
		dir := "projection-persistence"
		source, err := NewMediaFile("testdata/insta360.insp")
		require.NoError(t, err)
		originalName := filepath.Join(conf.OriginalsPath(), dir, "camera.insp")
		require.NoError(t, source.Copy(originalName, false))

		preview, err := NewMediaFile("testdata/insta360.insp.jpg")
		require.NoError(t, err)
		previewName := filepath.Join(conf.SidecarPath(), dir, "camera.insp.jpg")
		require.NoError(t, preview.Copy(previewName, false))

		generated, err := NewMediaFile(previewName)
		require.NoError(t, err)
		assert.Equal(t, projection.Equirectangular, generated.VisualProjection(""))
	})
	t.Run("OrdinarySidecar", func(t *testing.T) {
		conf := config.TestConfig()
		preview, err := NewMediaFile("testdata/flash.jpg")
		require.NoError(t, err)
		previewName := filepath.Join(conf.SidecarPath(), "ordinary.jpg")
		require.NoError(t, preview.Copy(previewName, false))

		generated, err := NewMediaFile(previewName)
		require.NoError(t, err)
		assert.Equal(t, projection.Unknown, generated.VisualProjection(""))
	})
}
