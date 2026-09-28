package photoprism

import (
	"os"
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
	t.Run("CaptureSidecars", func(t *testing.T) {
		conf := config.TestConfig()
		dir := "projection-capture"
		t.Cleanup(func() {
			_ = os.RemoveAll(filepath.Join(conf.OriginalsPath(), dir))
			_ = os.RemoveAll(filepath.Join(conf.SidecarPath(), dir))
		})

		originals := filepath.Join(conf.OriginalsPath(), dir)
		sidecars := filepath.Join(conf.SidecarPath(), dir)
		writeInsta360CaptureFile(t, originals, "VID_20220625_140410_00_008.insv", "testdata/flash.jpg")
		writeInsta360CaptureFile(t, originals, "VID_20220625_140410_10_008.insv", "testdata/flash.jpg")
		writeInsta360CaptureFile(t, originals, "LRV_20220625_140410_11_008.insv", "testdata/flash.jpg")

		// Every preview carries equirectangular metadata, as a single lens dewarped as both lenses does;
		// the proxy holds both lenses, so its dewarped preview keeps it.
		for name, expected := range map[string]projection.Type{
			"VID_20220625_140410_00_008.insv.jpg": projection.Equirectangular,
			"VID_20220625_140410_10_008.insv.jpg": projection.Unknown,
			"LRV_20220625_140410_11_008.insv.jpg": projection.Equirectangular,
		} {
			generated, err := NewMediaFile(writeInsta360CaptureFile(t, sidecars, name, "testdata/flash.jpg"))
			require.NoError(t, err)
			assert.Equal(t, expected, generated.VisualProjection(projection.Equirectangular.String()), name)
		}
	})
	t.Run("DualStreamSidecar", func(t *testing.T) {
		conf := config.TestConfig()
		dir := "projection-dual-stream"
		t.Cleanup(func() {
			_ = os.RemoveAll(filepath.Join(conf.OriginalsPath(), dir))
			_ = os.RemoveAll(filepath.Join(conf.SidecarPath(), dir))
		})

		newInsta360StreamFile(t, filepath.Join(conf.OriginalsPath(), dir), "clip.insv")
		preview, err := NewMediaFile(writeInsta360CaptureFile(t, filepath.Join(conf.SidecarPath(), dir), "clip.insv.jpg", "testdata/insta360.insp.jpg"))
		require.NoError(t, err)
		assert.Equal(t, projection.Equirectangular, preview.VisualProjection(""))
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

// TestMediaFile_GeneratedSourceName verifies that only generated JPEG and AVC sidecars map to an original.
func TestMediaFile_GeneratedSourceName(t *testing.T) {
	conf := config.TestConfig()
	dir := "generated-source"
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Join(conf.SidecarPath(), dir)) })

	for name, expected := range map[string]string{
		"camera.insp.jpg": filepath.Join(conf.OriginalsPath(), dir, "camera.insp"),
		"clip.insv.avc":   filepath.Join(conf.OriginalsPath(), dir, "clip.insv"),
		"camera.insp.png": "",
	} {
		sidecar, err := NewMediaFile(writeInsta360CaptureFile(t, filepath.Join(conf.SidecarPath(), dir), name, "testdata/flash.jpg"))
		require.NoError(t, err)
		assert.Equal(t, expected, sidecar.generatedSourceName(), name)
	}

	original, err := NewMediaFile("testdata/flash.jpg")
	require.NoError(t, err)
	assert.Equal(t, "", original.generatedSourceName())
	assert.Equal(t, "", (*MediaFile)(nil).generatedSourceName())
}
