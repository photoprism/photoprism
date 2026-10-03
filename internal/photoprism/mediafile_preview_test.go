package photoprism

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// newPreviewFolders creates a new originals folder and its sidecar folder, and returns both.
func newPreviewFolders(t *testing.T) (dir, sidecarDir string) {
	t.Helper()

	folder := "preview-lookup-" + rnd.Base36(8)
	dir = filepath.Join(Config().OriginalsPath(), folder)
	sidecarDir = filepath.Join(Config().SidecarPath(), folder)
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
		_ = os.RemoveAll(sidecarDir)
	})
	require.NoError(t, fs.MkdirAll(dir))
	require.NoError(t, fs.MkdirAll(sidecarDir))

	return dir, sidecarDir
}

// newMislabeledHeic writes a HEIC original with a mislabeled IMG_1.jpg next to it and a valid sidecar
// preview, and returns the HEIC and the name of the sidecar preview.
func newMislabeledHeic(t *testing.T) (*MediaFile, string) {
	t.Helper()

	dir, sidecarDir := newPreviewFolders(t)
	require.NoError(t, fs.Copy(filepath.Join(Config().SamplesPath(), "iphone_7.heic"), filepath.Join(dir, "IMG_1.heic"), false))
	require.NoError(t, fs.Copy("testdata/photoprism.png", filepath.Join(dir, "IMG_1.jpg"), false))
	sidecarName := filepath.Join(sidecarDir, "IMG_1.heic.jpg")
	require.NoError(t, fs.Copy("testdata/flash.jpg", sidecarName, false))

	f, err := NewMediaFile(filepath.Join(dir, "IMG_1.heic"))
	require.NoError(t, err)

	return f, sidecarName
}

func TestFindPreviewImage(t *testing.T) {
	find := func(fileName string, types ...fs.Type) *MediaFile {
		return findPreviewImage(fileName, Config().SidecarPath(), Config().OriginalsPath(), false, types...)
	}

	t.Run("SkipsMislabeled", func(t *testing.T) {
		f, sidecarName := newMislabeledHeic(t)
		preview := find(f.FileName(), fs.ImageJpeg, fs.ImagePng)
		require.NotNil(t, preview)
		assert.Equal(t, sidecarName, preview.FileName())
	})
	t.Run("OwnFolderFirst", func(t *testing.T) {
		dir, sidecarDir := newPreviewFolders(t)
		require.NoError(t, fs.Copy("testdata/flash.jpg", filepath.Join(dir, "IMG_1.jpg"), false))
		require.NoError(t, fs.Copy("testdata/flash.jpg", filepath.Join(sidecarDir, "IMG_1.heic.jpg"), false))
		preview := find(filepath.Join(dir, "IMG_1.heic"), fs.ImageJpeg, fs.ImagePng)
		require.NotNil(t, preview)
		assert.Equal(t, filepath.Join(dir, "IMG_1.jpg"), preview.FileName())
	})
	t.Run("TypeOrder", func(t *testing.T) {
		dir, sidecarDir := newPreviewFolders(t)
		require.NoError(t, fs.Copy("testdata/flash.jpg", filepath.Join(sidecarDir, "IMG_1.heic.jpg"), false))
		require.NoError(t, fs.Copy("testdata/photoprism.png", filepath.Join(sidecarDir, "IMG_1.heic.png"), false))
		fileName := filepath.Join(dir, "IMG_1.heic")
		assert.Equal(t, filepath.Join(sidecarDir, "IMG_1.heic.jpg"), find(fileName, fs.ImageJpeg, fs.ImagePng).FileName())
		assert.Equal(t, filepath.Join(sidecarDir, "IMG_1.heic.png"), find(fileName, fs.ImagePng, fs.ImageJpeg).FileName())
	})
	t.Run("NotFound", func(t *testing.T) {
		dir, _ := newPreviewFolders(t)
		require.NoError(t, fs.Copy("testdata/photoprism.png", filepath.Join(dir, "IMG_1.jpg"), false))
		assert.Nil(t, find(filepath.Join(dir, "IMG_1.heic"), fs.ImageJpeg, fs.ImagePng))
		assert.Nil(t, find("", fs.ImageJpeg))
	})
}

func TestPreviewContentMatches(t *testing.T) {
	t.Run("Jpeg", func(t *testing.T) {
		assert.True(t, previewContentMatches("testdata/flash.jpg", fs.ImageJpeg))
		assert.False(t, previewContentMatches("testdata/photoprism.png", fs.ImageJpeg))
	})
	t.Run("Png", func(t *testing.T) {
		assert.True(t, previewContentMatches("testdata/photoprism.png", fs.ImagePng))
		assert.False(t, previewContentMatches("testdata/flash.jpg", fs.ImagePng))
	})
	t.Run("Video", func(t *testing.T) {
		assert.False(t, previewContentMatches(filepath.Join(Config().SamplesPath(), "blue-go-video.mp4"), fs.ImageJpeg))
	})
	t.Run("OtherType", func(t *testing.T) {
		assert.False(t, previewContentMatches("testdata/flash.jpg", fs.ImageHeic))
		assert.False(t, previewContentMatches("testdata/missing.jpg", fs.ImageJpeg))
	})
}

func TestMediaFile_PreviewLookupMislabeled(t *testing.T) {
	t.Run("NoValidPreview", func(t *testing.T) {
		dir, _ := newPreviewFolders(t)
		require.NoError(t, fs.Copy(filepath.Join(Config().SamplesPath(), "iphone_7.heic"), filepath.Join(dir, "IMG_1.heic"), false))
		require.NoError(t, fs.Copy("testdata/photoprism.png", filepath.Join(dir, "IMG_1.jpg"), false))
		f, err := NewMediaFile(filepath.Join(dir, "IMG_1.heic"))
		require.NoError(t, err)
		assert.False(t, f.HasPreviewImage())
		preview, err := f.PreviewImage()
		assert.Error(t, err)
		assert.Nil(t, preview)
	})
	t.Run("HasPreviewImage", func(t *testing.T) {
		f, _ := newMislabeledHeic(t)
		assert.True(t, f.HasPreviewImage())
	})
	t.Run("PreviewImage", func(t *testing.T) {
		f, sidecarName := newMislabeledHeic(t)
		preview, err := f.PreviewImage()
		require.NoError(t, err)
		assert.Equal(t, sidecarName, preview.FileName())
	})
	t.Run("ToImage", func(t *testing.T) {
		// Converting again would rewrite the preview through the native HEIC decoder of libvips.
		f, sidecarName := newMislabeledHeic(t)
		past := time.Now().Add(-time.Hour).Truncate(time.Second)
		require.NoError(t, os.Chtimes(sidecarName, past, past))

		preview, err := NewConvert(Config()).ToImage(f, false)
		require.NoError(t, err)
		assert.Equal(t, sidecarName, preview.FileName())
		info, err := os.Stat(sidecarName)
		require.NoError(t, err)
		assert.True(t, info.ModTime().Equal(past), "the existing preview is not converted again")
	})
}
