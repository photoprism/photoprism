package commands

import (
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestVideoReindexRelated(t *testing.T) {
	t.Run("MissingFileName", func(t *testing.T) {
		assert.Error(t, videoReindexRelated(get.Config(), ""))
	})
	t.Run("UnportableSibling", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		conf := get.Config()
		dir := "video-reindex-" + rnd.Base36(8)
		root := filepath.Join(conf.OriginalsPath(), dir)
		require.NoError(t, fs.MkdirAll(root))
		t.Cleanup(func() {
			_ = os.RemoveAll(root)
			var photoIDs []uint
			db := entity.UnscopedDb()
			_ = db.Model(&entity.File{}).Where("file_root = ? AND file_name LIKE ?", entity.RootOriginals, dir+"/%").Pluck("photo_id", &photoIDs).Error
			_ = db.Where("file_root = ? AND file_name LIKE ?", entity.RootOriginals, dir+"/%").Delete(&entity.File{}).Error
			if len(photoIDs) > 0 {
				_ = db.Where("id IN (?)", photoIDs).Delete(&entity.Photo{}).Error
			}
			_ = db.Where("root = ? AND path = ?", entity.RootOriginals, dir).Delete(&entity.Folder{}).Error
		})

		jpegFile, err := os.Create(filepath.Join(root, "IMG.jpg")) //nolint:gosec // test-owned path
		require.NoError(t, err)
		require.NoError(t, jpeg.Encode(jpegFile, image.NewRGBA(image.Rect(0, 0, 16, 16)), nil))
		require.NoError(t, jpegFile.Close())

		gifFile, err := os.Create(filepath.Join(root, "IMG.x|y.gif")) //nolint:gosec // test-owned path
		require.NoError(t, err)
		require.NoError(t, gif.Encode(gifFile, image.NewPaletted(image.Rect(0, 0, 16, 16), color.Palette{color.Black, color.White}), nil))
		require.NoError(t, gifFile.Close())

		require.NoError(t, videoReindexRelated(conf, filepath.Join(root, "IMG.jpg")))

		// indexed reports whether a file below the test folder has an index entry.
		indexed := func(name string) bool {
			var count int64
			require.NoError(t, entity.UnscopedDb().Model(&entity.File{}).
				Where("file_root = ? AND file_name = ?", entity.RootOriginals, dir+"/"+name).Count(&count).Error)
			return count > 0
		}

		assert.True(t, indexed("IMG.jpg"))
		assert.False(t, indexed("IMG.x|y.gif"))
	})
}
