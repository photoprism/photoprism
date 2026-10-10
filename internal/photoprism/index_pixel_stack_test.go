package photoprism

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// newGooglePixelStackConfig returns an isolated config with its own database for Google Pixel Camera testing.
func newGooglePixelStackConfig(t *testing.T, dbName string) *config.Config {
	t.Helper()

	cfg := config.NewMinimalTestConfigWithDb(dbName, filepath.Join(t.TempDir(), "storage"))
	oldCfg := Config()
	SetConfig(cfg)
	t.Cleanup(func() {
		SetConfig(oldCfg)
		oldCfg.RegisterDb()
	})

	return cfg
}

// writeGooglePixelZeroGpsJpeg generates a simple JPEG image without any EXIF or GPS metadata.
func writeGooglePixelZeroGpsJpeg(t *testing.T, fileName string, tint ...uint8) {
	t.Helper()

	require.NoError(t, fs.MkdirAll(filepath.Dir(fileName)))

	b := uint8(100)
	if len(tint) > 0 {
		b = tint[0]
	}

	img := image.NewRGBA(image.Rect(0, 0, 160, 120))
	for y := 0; y < 120; y++ {
		for x := 0; x < 160; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: b, A: 255})
		}
	}

	f, err := os.Create(fileName)
	require.NoError(t, err)
	defer f.Close()

	require.NoError(t, jpeg.Encode(f, img, &jpeg.Options{Quality: 85}))
}

// writeGooglePixelZeroGpsDng copies canon_eos_6d.dng, which has no GPS coordinates.
func writeGooglePixelZeroGpsDng(t *testing.T, cfg *config.Config, fileName string) {
	t.Helper()

	require.NoError(t, fs.MkdirAll(filepath.Dir(fileName)))
	src := filepath.Join(cfg.SamplesPath(), "canon_eos_6d.dng")
	require.NoError(t, fs.Copy(src, fileName, false))
}

// writeGooglePixelVideo writes a video file by copying source and appending unique identifier bytes.
func writeGooglePixelVideo(t *testing.T, src, dest, identifier string) {
	t.Helper()

	require.NoError(t, fs.MkdirAll(filepath.Dir(dest)))
	require.NoError(t, fs.Copy(src, dest, false))

	f, err := os.OpenFile(dest, os.O_APPEND|os.O_WRONLY, 0)
	require.NoError(t, err)
	_, err = f.WriteString(identifier)
	require.NoError(t, err)
	require.NoError(t, f.Close())
}

// indexGooglePixelFolder indexes a folder below the originals path.
func indexGooglePixelFolder(cfg *config.Config, folder string, rescan bool) {
	ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())
	ind.Start(NewIndexOptions(folder, rescan, true, true, false, false, cfg))
}

// googlePixelStackOwners returns the photos that own the originals in folder, keyed by file name.
func googlePixelStackOwners(t *testing.T, folder string) map[string]entity.Photo {
	t.Helper()

	var files []entity.File
	require.NoError(t, entity.UnscopedDb().
		Where("file_root = ? AND file_name LIKE ?", entity.RootOriginals, folder+"/%").
		Find(&files).Error)

	result := make(map[string]entity.Photo, len(files))

	for _, f := range files {
		var p entity.Photo
		require.NoError(t, entity.UnscopedDb().First(&p, "id = ?", f.PhotoID).Error)
		result[filepath.Base(f.FileName)] = p
	}

	return result
}

// googlePixelStackPreviews returns the previews in folder, keyed by file name.
func googlePixelStackPreviews(t *testing.T, folder string) map[string]entity.File {
	t.Helper()

	var files []entity.File
	require.NoError(t, entity.UnscopedDb().
		Where("file_root = ? AND file_name LIKE ?", entity.RootSidecar, folder+"/%").
		Find(&files).Error)

	result := make(map[string]entity.File, len(files))
	for _, f := range files {
		result[filepath.Base(f.FileName)] = f
	}

	return result
}

// TestIndex_GooglePixel_Simultaneous verifies that simultaneous indexing of a Google Pixel Camera capture with zero GPS metadata
// creates exactly one photo with Cover as primary.
func TestIndex_GooglePixel_Simultaneous(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// Standard Photo RAW+JPEG: verifies that modern capture (without .COVER) stacks into 1 photo with Cover as primary.
	t.Run("Modern", func(t *testing.T) {
		folder := "pixel_simultaneous_modern"
		cfg := newGooglePixelStackConfig(t, folder)
		dir := filepath.Join(cfg.OriginalsPath(), folder)

		coverPath := filepath.Join(dir, "PXL_20240926_143000123.RAW-01.jpg")
		rawPath := filepath.Join(dir, "PXL_20240926_143000123.RAW-02.ORIGINAL.dng")

		writeGooglePixelZeroGpsJpeg(t, coverPath)
		writeGooglePixelZeroGpsDng(t, cfg, rawPath)

		indexGooglePixelFolder(cfg, folder, false)

		var photos []entity.Photo
		require.NoError(t, entity.UnscopedDb().Where("photo_path = ?", folder).Find(&photos).Error)
		assert.Len(t, photos, 1, "expected exactly 1 photo for Google Pixel Camera capture pair")

		if len(photos) == 1 {
			photo := photos[0]
			assert.Equal(t, "PXL_20240926_143000123", photo.PhotoName)

			var files []entity.File
			require.NoError(t, entity.UnscopedDb().Where("photo_id = ?", photo.ID).Find(&files).Error)
			assert.Len(t, files, 2, "expected exactly 2 files stacked under the photo")

			var coverFile, rawFile *entity.File
			for i := range files {
				if strings.HasSuffix(files[i].FileName, "RAW-01.jpg") {
					coverFile = &files[i]
				} else if strings.HasSuffix(files[i].FileName, "ORIGINAL.dng") {
					rawFile = &files[i]
				}
			}

			require.NotNil(t, coverFile, "cover file must exist")
			require.NotNil(t, rawFile, "raw file must exist")
			assert.True(t, coverFile.FilePrimary, "cover file must be primary")
			assert.False(t, rawFile.FilePrimary, "raw file must not be primary")
		}
	})

	// Standard Photo RAW+JPEG (Legacy .COVER): verifies that legacy capture with .COVER stacks into 1 photo with Cover as primary.
	t.Run("Legacy", func(t *testing.T) {
		folder := "pixel_simultaneous_legacy"
		cfg := newGooglePixelStackConfig(t, folder)
		dir := filepath.Join(cfg.OriginalsPath(), folder)

		coverPath := filepath.Join(dir, "PXL_20240926_143000123.RAW-01.COVER.jpg")
		rawPath := filepath.Join(dir, "PXL_20240926_143000123.RAW-02.ORIGINAL.dng")

		writeGooglePixelZeroGpsJpeg(t, coverPath)
		writeGooglePixelZeroGpsDng(t, cfg, rawPath)

		indexGooglePixelFolder(cfg, folder, false)

		var photos []entity.Photo
		require.NoError(t, entity.UnscopedDb().Where("photo_path = ?", folder).Find(&photos).Error)
		assert.Len(t, photos, 1, "expected exactly 1 photo for Google Pixel Camera capture pair")

		if len(photos) == 1 {
			photo := photos[0]
			assert.Equal(t, "PXL_20240926_143000123", photo.PhotoName)

			var files []entity.File
			require.NoError(t, entity.UnscopedDb().Where("photo_id = ?", photo.ID).Find(&files).Error)
			assert.Len(t, files, 2, "expected exactly 2 files stacked under the photo")

			var coverFile, rawFile *entity.File
			for i := range files {
				if strings.HasSuffix(files[i].FileName, "COVER.jpg") {
					coverFile = &files[i]
				} else if strings.HasSuffix(files[i].FileName, "ORIGINAL.dng") {
					rawFile = &files[i]
				}
			}

			require.NotNil(t, coverFile, "cover file must exist")
			require.NotNil(t, rawFile, "raw file must exist")
			assert.True(t, coverFile.FilePrimary, "cover file must be primary")
			assert.False(t, rawFile.FilePrimary, "raw file must not be primary")
		}
	})
}

// TestIndex_GooglePixel_Portrait_Simultaneous verifies stacking for Google Pixel Camera Portrait mode pairs.
func TestIndex_GooglePixel_Portrait_Simultaneous(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// Portrait Photo JPEG: verifies that modern blurred portrait photo is primary over the unblurred original.
	t.Run("Modern", func(t *testing.T) {
		folder := "pixel_portrait_modern"
		cfg := newGooglePixelStackConfig(t, folder)
		dir := filepath.Join(cfg.OriginalsPath(), folder)

		coverPath := filepath.Join(dir, "PXL_20240926_150000123.PORTRAIT.jpg")
		origPath := filepath.Join(dir, "PXL_20240926_150000123.PORTRAIT.ORIGINAL.jpg")

		writeGooglePixelZeroGpsJpeg(t, coverPath, 50)
		writeGooglePixelZeroGpsJpeg(t, origPath, 150)

		indexGooglePixelFolder(cfg, folder, false)

		var photos []entity.Photo
		require.NoError(t, entity.UnscopedDb().Where("photo_path = ?", folder).Find(&photos).Error)
		assert.Len(t, photos, 1, "expected exactly 1 photo for Portrait pair")

		if len(photos) == 1 {
			photo := photos[0]
			assert.Equal(t, "PXL_20240926_150000123", photo.PhotoName)

			var files []entity.File
			require.NoError(t, entity.UnscopedDb().Where("photo_id = ?", photo.ID).Find(&files).Error)
			assert.Len(t, files, 2)

			var coverFile, origFile *entity.File
			for i := range files {
				if strings.HasSuffix(files[i].FileName, "PORTRAIT.ORIGINAL.jpg") {
					origFile = &files[i]
				} else if strings.HasSuffix(files[i].FileName, "PORTRAIT.jpg") {
					coverFile = &files[i]
				}
			}
			require.NotNil(t, coverFile, "portrait cover file should be linked to photo")
			require.NotNil(t, origFile, "portrait original file should be linked to photo")
			assert.True(t, coverFile.FilePrimary, "blurred portrait photo must be primary")
			assert.False(t, origFile.FilePrimary, "unblurred original photo must not be primary")
		}
	})

	// Portrait Photo JPEG (legacy -01/-02): verifies that legacy blurred cover photo is primary over the unblurred original.
	t.Run("Legacy", func(t *testing.T) {
		folder := "pixel_portrait_legacy"
		cfg := newGooglePixelStackConfig(t, folder)
		dir := filepath.Join(cfg.OriginalsPath(), folder)

		coverPath := filepath.Join(dir, "PXL_20240926_150000123.PORTRAIT-01.COVER.jpg")
		origPath := filepath.Join(dir, "PXL_20240926_150000123.PORTRAIT-02.ORIGINAL.jpg")

		writeGooglePixelZeroGpsJpeg(t, coverPath, 50)
		writeGooglePixelZeroGpsJpeg(t, origPath, 150)

		indexGooglePixelFolder(cfg, folder, false)

		var photos []entity.Photo
		require.NoError(t, entity.UnscopedDb().Where("photo_path = ?", folder).Find(&photos).Error)
		assert.Len(t, photos, 1, "expected exactly 1 photo for Legacy Portrait pair")

		if len(photos) == 1 {
			photo := photos[0]
			assert.Equal(t, "PXL_20240926_150000123", photo.PhotoName)

			var files []entity.File
			require.NoError(t, entity.UnscopedDb().Where("photo_id = ?", photo.ID).Find(&files).Error)
			assert.Len(t, files, 2)

			var coverFile, origFile *entity.File
			for i := range files {
				if strings.HasSuffix(files[i].FileName, "PORTRAIT-02.ORIGINAL.jpg") {
					origFile = &files[i]
				} else if strings.HasSuffix(files[i].FileName, "PORTRAIT-01.COVER.jpg") {
					coverFile = &files[i]
				}
			}
			require.NotNil(t, coverFile, "portrait cover file should be linked to photo")
			require.NotNil(t, origFile, "portrait original file should be linked to photo")
			assert.True(t, coverFile.FilePrimary, "blurred portrait cover photo must be primary")
			assert.False(t, origFile.FilePrimary, "unblurred original photo must not be primary")
		}
	})
}

// TestIndex_GooglePixel_Sequential_LateCover verifies that when the Google Pixel Camera RAW file is indexed first,
// a late-arriving Cover image stacks into the existing photo and is promoted to FilePrimary.
func TestIndex_GooglePixel_Sequential_LateCover(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// Standard Photo RAW+JPEG: verifies that late-arriving modern Cover JPEG is promoted to FilePrimary.
	t.Run("Modern", func(t *testing.T) {
		folder := "pixel_sequential_modern"
		cfg := newGooglePixelStackConfig(t, folder)
		dir := filepath.Join(cfg.OriginalsPath(), folder)

		rawPath := filepath.Join(dir, "PXL_20240926_160000123.RAW-02.ORIGINAL.dng")
		writeGooglePixelZeroGpsDng(t, cfg, rawPath)

		// Step 1: Index RAW first.
		indexGooglePixelFolder(cfg, folder, false)

		var photos []entity.Photo
		require.NoError(t, entity.UnscopedDb().Where("photo_path = ?", folder).Find(&photos).Error)
		require.Len(t, photos, 1, "initial photo for RAW file must exist")
		initialPhoto := photos[0]

		// Step 2: Write Cover image and re-index.
		coverPath := filepath.Join(dir, "PXL_20240926_160000123.RAW-01.jpg")
		writeGooglePixelZeroGpsJpeg(t, coverPath)

		indexGooglePixelFolder(cfg, folder, false)

		var photosAfter []entity.Photo
		require.NoError(t, entity.UnscopedDb().Where("photo_path = ? AND deleted_at IS NULL", folder).Find(&photosAfter).Error)
		assert.Len(t, photosAfter, 1, "should still have only 1 active photo after late Cover arrives")

		var files []entity.File
		require.NoError(t, entity.UnscopedDb().Where("photo_id = ?", initialPhoto.ID).Find(&files).Error)
		assert.GreaterOrEqual(t, len(files), 2, "cover and raw files should be linked to initial photo")

		var coverFile *entity.File
		for i := range files {
			if strings.HasSuffix(files[i].FileName, "RAW-01.jpg") {
				coverFile = &files[i]
			}
		}
		require.NotNil(t, coverFile, "cover file should be linked to photo")
		assert.True(t, coverFile.FilePrimary, "late Cover file must be promoted to FilePrimary")
	})

	// Standard Photo RAW+JPEG (Legacy .COVER): verifies that late-arriving legacy Cover JPEG is promoted to FilePrimary.
	t.Run("Legacy", func(t *testing.T) {
		folder := "pixel_sequential_legacy"
		cfg := newGooglePixelStackConfig(t, folder)
		dir := filepath.Join(cfg.OriginalsPath(), folder)

		rawPath := filepath.Join(dir, "PXL_20240926_160000123.RAW-02.ORIGINAL.dng")
		writeGooglePixelZeroGpsDng(t, cfg, rawPath)

		// Step 1: Index RAW first.
		indexGooglePixelFolder(cfg, folder, false)

		var photos []entity.Photo
		require.NoError(t, entity.UnscopedDb().Where("photo_path = ?", folder).Find(&photos).Error)
		require.Len(t, photos, 1, "initial photo for RAW file must exist")
		initialPhoto := photos[0]

		// Step 2: Write Cover image and re-index.
		coverPath := filepath.Join(dir, "PXL_20240926_160000123.RAW-01.COVER.jpg")
		writeGooglePixelZeroGpsJpeg(t, coverPath)

		indexGooglePixelFolder(cfg, folder, false)

		var photosAfter []entity.Photo
		require.NoError(t, entity.UnscopedDb().Where("photo_path = ? AND deleted_at IS NULL", folder).Find(&photosAfter).Error)
		assert.Len(t, photosAfter, 1, "should still have only 1 active photo after late Cover arrives")

		var files []entity.File
		require.NoError(t, entity.UnscopedDb().Where("photo_id = ?", initialPhoto.ID).Find(&files).Error)
		assert.GreaterOrEqual(t, len(files), 2, "cover and raw files should be linked to initial photo")

		var coverFile *entity.File
		for i := range files {
			if strings.HasSuffix(files[i].FileName, "COVER.jpg") {
				coverFile = &files[i]
			}
		}
		require.NotNil(t, coverFile, "cover file should be linked to photo")
		assert.True(t, coverFile.FilePrimary, "late Cover file must be promoted to FilePrimary")
	})
}

// TestIndex_GooglePixel_Sequential_Portrait_LateCover verifies that when a Google Pixel Camera Portrait Original JPEG is indexed first,
// a late-arriving Portrait Cover image stacks into the existing photo and is promoted to FilePrimary.
func TestIndex_GooglePixel_Sequential_Portrait_LateCover(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// Portrait Photo JPEG: verifies that late-arriving modern blurred Cover is promoted over original JPEG.
	t.Run("Modern", func(t *testing.T) {
		folder := "pixel_portrait_sequential_modern"
		cfg := newGooglePixelStackConfig(t, folder)
		dir := filepath.Join(cfg.OriginalsPath(), folder)

		origPath := filepath.Join(dir, "PXL_20240926_180000123.PORTRAIT.ORIGINAL.jpg")
		writeGooglePixelZeroGpsJpeg(t, origPath, 150)

		// Step 1: Index Original JPEG first.
		indexGooglePixelFolder(cfg, folder, false)

		var photos []entity.Photo
		require.NoError(t, entity.UnscopedDb().Where("photo_path = ?", folder).Find(&photos).Error)
		require.Len(t, photos, 1, "initial photo for Original JPEG file must exist")
		initialPhoto := photos[0]

		// Step 2: Write Cover JPEG image and re-index.
		coverPath := filepath.Join(dir, "PXL_20240926_180000123.PORTRAIT.jpg")
		writeGooglePixelZeroGpsJpeg(t, coverPath, 50)

		indexGooglePixelFolder(cfg, folder, false)

		var photosAfter []entity.Photo
		require.NoError(t, entity.UnscopedDb().Where("photo_path = ? AND deleted_at IS NULL", folder).Find(&photosAfter).Error)
		assert.Len(t, photosAfter, 1, "should still have only 1 active photo after late Cover arrives")

		var files []entity.File
		require.NoError(t, entity.UnscopedDb().Where("photo_id = ?", initialPhoto.ID).Find(&files).Error)
		assert.Len(t, files, 2, "cover and original files should be linked to initial photo")

		var coverFile, origFile *entity.File
		for i := range files {
			if strings.HasSuffix(files[i].FileName, "PORTRAIT.jpg") {
				coverFile = &files[i]
			} else if strings.HasSuffix(files[i].FileName, "PORTRAIT.ORIGINAL.jpg") {
				origFile = &files[i]
			}
		}
		require.NotNil(t, coverFile, "cover file should be linked to photo")
		require.NotNil(t, origFile, "original file should be linked to photo")
		assert.True(t, coverFile.FilePrimary, "late Cover file must be promoted to FilePrimary over companion original")
		assert.False(t, origFile.FilePrimary, "original file must be demoted from FilePrimary")
	})

	// Portrait Photo JPEG (legacy -01/-02): verifies that late-arriving legacy blurred Cover is promoted over original JPEG.
	t.Run("Legacy", func(t *testing.T) {
		folder := "pixel_portrait_sequential_legacy"
		cfg := newGooglePixelStackConfig(t, folder)
		dir := filepath.Join(cfg.OriginalsPath(), folder)

		origPath := filepath.Join(dir, "PXL_20240926_180000123.PORTRAIT-02.ORIGINAL.jpg")
		writeGooglePixelZeroGpsJpeg(t, origPath, 150)

		// Step 1: Index Original JPEG first.
		indexGooglePixelFolder(cfg, folder, false)

		var photos []entity.Photo
		require.NoError(t, entity.UnscopedDb().Where("photo_path = ?", folder).Find(&photos).Error)
		require.Len(t, photos, 1, "initial photo for Original JPEG file must exist")
		initialPhoto := photos[0]

		// Step 2: Write Cover JPEG image and re-index.
		coverPath := filepath.Join(dir, "PXL_20240926_180000123.PORTRAIT-01.COVER.jpg")
		writeGooglePixelZeroGpsJpeg(t, coverPath, 50)

		indexGooglePixelFolder(cfg, folder, false)

		var photosAfter []entity.Photo
		require.NoError(t, entity.UnscopedDb().Where("photo_path = ? AND deleted_at IS NULL", folder).Find(&photosAfter).Error)
		assert.Len(t, photosAfter, 1, "should still have only 1 active photo after late Cover arrives")

		var files []entity.File
		require.NoError(t, entity.UnscopedDb().Where("photo_id = ?", initialPhoto.ID).Find(&files).Error)
		assert.Len(t, files, 2, "cover and original files should be linked to initial photo")

		var coverFile, origFile *entity.File
		for i := range files {
			if strings.HasSuffix(files[i].FileName, "COVER.jpg") {
				coverFile = &files[i]
			} else if strings.HasSuffix(files[i].FileName, "ORIGINAL.jpg") {
				origFile = &files[i]
			}
		}
		require.NotNil(t, coverFile, "cover file should be linked to photo")
		require.NotNil(t, origFile, "original file should be linked to photo")
		assert.True(t, coverFile.FilePrimary, "late Cover file must be promoted to FilePrimary over companion original")
		assert.False(t, origFile.FilePrimary, "original file must be demoted from FilePrimary")
	})
}

// TestIndex_GooglePixel_Rescan_Reconciliation verifies that forced reindexing reconciles duplicate photos
// created prior to Google Pixel Camera stacking support, migrating albums, labels, and keywords.
func TestIndex_GooglePixel_Rescan_Reconciliation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	folder := "pixel_reconcile"
	cfg := newGooglePixelStackConfig(t, folder)
	dir := filepath.Join(cfg.OriginalsPath(), folder)

	coverPath := filepath.Join(dir, "PXL_20240926_170000123.RAW-01.jpg")
	rawPath := filepath.Join(dir, "PXL_20240926_170000123.RAW-02.ORIGINAL.dng")
	writeGooglePixelZeroGpsJpeg(t, coverPath)
	writeGooglePixelZeroGpsDng(t, cfg, rawPath)

	coverMedia, err := NewMediaFile(coverPath)
	require.NoError(t, err)
	rawMedia, err := NewMediaFile(rawPath)
	require.NoError(t, err)

	// Seed DB with two separate photos simulating pre-fix state.
	photoCover := entity.NewPhoto(true)
	photoCover.PhotoPath = folder
	photoCover.PhotoName = "PXL_20240926_170000123.RAW-01"
	require.NoError(t, photoCover.Create())

	photoRaw := entity.NewPhoto(true)
	photoRaw.PhotoPath = folder
	photoRaw.PhotoName = "PXL_20240926_170000123.RAW-02.ORIGINAL"
	require.NoError(t, photoRaw.Create())

	fileCover := entity.File{
		PhotoID:     photoCover.ID,
		PhotoUID:    photoCover.PhotoUID,
		FileName:    coverMedia.RootRelName(),
		FileRoot:    entity.RootOriginals,
		FileHash:    coverMedia.Hash(),
		FileType:    coverMedia.FileType().String(),
		MediaType:   media.Image.String(),
		FilePrimary: true,
	}
	require.NoError(t, fileCover.Create())

	fileRaw := entity.File{
		PhotoID:     photoRaw.ID,
		PhotoUID:    photoRaw.PhotoUID,
		FileName:    rawMedia.RootRelName(),
		FileRoot:    entity.RootOriginals,
		FileHash:    rawMedia.Hash(),
		FileType:    rawMedia.FileType().String(),
		MediaType:   media.Image.String(),
		FilePrimary: true,
	}
	require.NoError(t, fileRaw.Create())

	// Add associations to photoRaw to verify migration.
	albumUID := rnd.GenerateUID(entity.AlbumUID)
	require.NoError(t, entity.NewPhotoAlbum(photoRaw.PhotoUID, albumUID).Create())
	require.NoError(t, entity.NewPhotoLabel(photoRaw.ID, 1234, 100, entity.SrcManual).Create())
	require.NoError(t, entity.NewPhotoKeyword(photoRaw.ID, 5678).Create())

	related, err := coverMedia.RelatedFiles(false)
	require.NoError(t, err)
	require.NoError(t, reconcileGooglePixelPhotos(related))

	// Verify photoRaw is soft-deleted.
	var photoRawAfter entity.Photo
	require.NoError(t, entity.UnscopedDb().First(&photoRawAfter, "id = ?", photoRaw.ID).Error)
	assert.True(t, photoRawAfter.DeletedAt != nil, "duplicate RAW photo must be soft-deleted")
	assert.Equal(t, -1, photoRawAfter.PhotoQuality)

	// Verify photoCover owns both files.
	var filesAfter []entity.File
	require.NoError(t, entity.UnscopedDb().Where("photo_id = ?", photoCover.ID).Find(&filesAfter).Error)
	assert.GreaterOrEqual(t, len(filesAfter), 2, "canonical cover photo must own both files")

	// Verify associations were migrated to photoCover.
	var albumCount, labelCount, keywordCount int
	require.NoError(t, entity.UnscopedDb().Model(&entity.PhotoAlbum{}).Where("photo_uid = ? AND album_uid = ?", photoCover.PhotoUID, albumUID).Count(&albumCount).Error)
	assert.Equal(t, 1, albumCount, "album association should be migrated to canonical photo")

	require.NoError(t, entity.UnscopedDb().Model(&entity.PhotoLabel{}).Where("photo_id = ? AND label_id = ?", photoCover.ID, 1234).Count(&labelCount).Error)
	assert.Equal(t, 1, labelCount, "label association should be migrated to canonical photo")

	require.NoError(t, entity.UnscopedDb().Model(&entity.PhotoKeyword{}).Where("photo_id = ? AND keyword_id = ?", photoCover.ID, 5678).Count(&keywordCount).Error)
	assert.Equal(t, 1, keywordCount, "keyword association should be migrated to canonical photo")

	// Verify that full folder rescan runs without error.
	indexGooglePixelFolder(cfg, folder, true)
}

// TestIndex_GooglePixel_AIProZoom_ThreeFiles verifies that a 3-file AI Pro Zoom capture (Cover JPEG, base JPEG, and RAW DNG)
// indexes as a single photo entity with the AI Pro Zoom Cover as FilePrimary.
func TestIndex_GooglePixel_AIProZoom_ThreeFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	folder := "pixel_aiprozoom"
	cfg := newGooglePixelStackConfig(t, folder)
	dir := filepath.Join(cfg.OriginalsPath(), folder)

	coverPath := filepath.Join(dir, "PXL_20240930_120000123.BURST-01.jpg")
	origJpgPath := filepath.Join(dir, "PXL_20240930_120000123.BURST-02.original.jpg")
	origDngPath := filepath.Join(dir, "PXL_20240930_120000123.BURST-03.ORIGINAL.dng")

	writeGooglePixelZeroGpsJpeg(t, coverPath, 100)
	writeGooglePixelZeroGpsJpeg(t, origJpgPath, 200)
	writeGooglePixelZeroGpsDng(t, cfg, origDngPath)

	indexGooglePixelFolder(cfg, folder, false)

	var photos []entity.Photo
	require.NoError(t, entity.UnscopedDb().Where("photo_path = ?", folder).Find(&photos).Error)
	assert.Len(t, photos, 1, "expected exactly 1 photo for Google Pixel Camera AI Pro Zoom 3-file capture")

	var files []entity.File
	require.NoError(t, entity.UnscopedDb().Where("photo_id = ?", photos[0].ID).Find(&files).Error)
	assert.Len(t, files, 3, "expected 3 files linked to the photo")

	var coverFile, origJpgFile, origDngFile *entity.File
	for i := range files {
		if strings.HasSuffix(files[i].FileName, "BURST-01.jpg") {
			coverFile = &files[i]
		} else if strings.HasSuffix(files[i].FileName, "BURST-02.original.jpg") {
			origJpgFile = &files[i]
		} else if strings.HasSuffix(files[i].FileName, "BURST-03.ORIGINAL.dng") {
			origDngFile = &files[i]
		}
	}

	require.NotNil(t, coverFile, "BURST-01 cover file should be linked to photo")
	require.NotNil(t, origJpgFile, "BURST-02 original file should be linked to photo")
	require.NotNil(t, origDngFile, "BURST-03 DNG file should be linked to photo")

	assert.True(t, coverFile.FilePrimary, "BURST-01 Cover must be FilePrimary")
	assert.False(t, origJpgFile.FilePrimary, "BURST-02 original must not be FilePrimary")
	assert.False(t, origDngFile.FilePrimary, "BURST-03 DNG must not be FilePrimary")
}

// TestIndex_GooglePixel_VideoBoost_Pair verifies that Video Boost and Night Sight Video clips
// (draft preview COVER and boosted MAIN) index as a single video entity with the boosted MAIN file as FilePrimary.
func TestIndex_GooglePixel_VideoBoost_Pair(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// Video Boost MP4: verifies that the boosted MAIN video is primary over the draft COVER preview.
	t.Run("VideoBoost", func(t *testing.T) {
		folder := "pixel_videoboost"
		cfg := newGooglePixelStackConfig(t, folder)
		dir := filepath.Join(cfg.OriginalsPath(), folder)

		videoSource := filepath.Join(cfg.SamplesPath(), "blue-go-video.mp4")
		coverVideo := filepath.Join(dir, "PXL_20240930_180000123.VB-01.COVER.mp4")
		mainVideo := filepath.Join(dir, "PXL_20240930_180000123.VB-03.MAIN.mp4")

		writeGooglePixelVideo(t, videoSource, coverVideo, "VB-01.COVER")
		writeGooglePixelVideo(t, videoSource, mainVideo, "VB-03.MAIN")

		indexGooglePixelFolder(cfg, folder, false)

		var photos []entity.Photo
		require.NoError(t, entity.UnscopedDb().Where("photo_path = ?", folder).Find(&photos).Error)
		assert.Len(t, photos, 1, "expected exactly 1 video entity for Video Boost pair")

		var files []entity.File
		require.NoError(t, entity.UnscopedDb().Where("photo_id = ? AND file_root = ?", photos[0].ID, entity.RootOriginals).Find(&files).Error)
		assert.Len(t, files, 2, "expected 2 files linked to the video entity")

		var coverFile, mainFile *entity.File
		for i := range files {
			if strings.HasSuffix(files[i].FileName, "VB-01.COVER.mp4") {
				coverFile = &files[i]
			} else if strings.HasSuffix(files[i].FileName, "VB-03.MAIN.mp4") {
				mainFile = &files[i]
			}
		}

		require.NotNil(t, coverFile, "VB-01.COVER video should be linked to photo")
		require.NotNil(t, mainFile, "VB-03.MAIN video should be linked to photo")

		assert.False(t, mainFile.FilePrimary, "video original file is never FilePrimary")
		assert.False(t, coverFile.FilePrimary, "draft COVER preview video must not be FilePrimary")

		previews := googlePixelStackPreviews(t, folder)
		require.Contains(t, previews, "PXL_20240930_180000123.VB-03.MAIN.mp4.jpg")
		require.Contains(t, previews, "PXL_20240930_180000123.VB-01.COVER.mp4.jpg")
		assert.True(t, previews["PXL_20240930_180000123.VB-03.MAIN.mp4.jpg"].FilePrimary, "boosted MAIN video preview must be FilePrimary")
		assert.False(t, previews["PXL_20240930_180000123.VB-01.COVER.mp4.jpg"].FilePrimary, "draft COVER preview must not be FilePrimary")
	})

	// Night Sight Video MP4: verifies that the boosted MAIN video is primary over the draft COVER preview.
	t.Run("NightSightVideo", func(t *testing.T) {
		folder := "pixel_nightsightvideo"
		cfg := newGooglePixelStackConfig(t, folder)
		dir := filepath.Join(cfg.OriginalsPath(), folder)

		videoSource := filepath.Join(cfg.SamplesPath(), "blue-go-video.mp4")
		coverVideo := filepath.Join(dir, "PXL_20260930_200000123.NS-01.COVER.mp4")
		mainVideo := filepath.Join(dir, "PXL_20260930_200000123.NS-02.MAIN.mp4")

		writeGooglePixelVideo(t, videoSource, coverVideo, "NS-01.COVER")
		writeGooglePixelVideo(t, videoSource, mainVideo, "NS-02.MAIN")

		indexGooglePixelFolder(cfg, folder, false)

		var photos []entity.Photo
		require.NoError(t, entity.UnscopedDb().Where("photo_path = ?", folder).Find(&photos).Error)
		assert.Len(t, photos, 1, "expected exactly 1 video entity for Night Sight Video pair")

		var files []entity.File
		require.NoError(t, entity.UnscopedDb().Where("photo_id = ? AND file_root = ?", photos[0].ID, entity.RootOriginals).Find(&files).Error)
		assert.Len(t, files, 2, "expected 2 files linked to the video entity")

		var coverFile, mainFile *entity.File
		for i := range files {
			if strings.HasSuffix(files[i].FileName, "NS-01.COVER.mp4") {
				coverFile = &files[i]
			} else if strings.HasSuffix(files[i].FileName, "NS-02.MAIN.mp4") {
				mainFile = &files[i]
			}
		}

		require.NotNil(t, coverFile, "NS-01.COVER video should be linked to photo")
		require.NotNil(t, mainFile, "NS-02.MAIN video should be linked to photo")

		assert.False(t, mainFile.FilePrimary, "video original file is never FilePrimary")
		assert.False(t, coverFile.FilePrimary, "draft COVER preview video must not be FilePrimary")

		previews := googlePixelStackPreviews(t, folder)
		require.Contains(t, previews, "PXL_20260930_200000123.NS-02.MAIN.mp4.jpg")
		require.Contains(t, previews, "PXL_20260930_200000123.NS-01.COVER.mp4.jpg")
		assert.True(t, previews["PXL_20260930_200000123.NS-02.MAIN.mp4.jpg"].FilePrimary, "boosted MAIN video preview must be FilePrimary")
		assert.False(t, previews["PXL_20260930_200000123.NS-01.COVER.mp4.jpg"].FilePrimary, "draft COVER preview must not be FilePrimary")
	})
}

// TestIndex_GooglePixel_Sequential_VideoBoost_LateMain verifies that when the draft preview COVER video
// is indexed first, a late-arriving boosted MAIN video stacks into the existing photo, is promoted to
// FilePrimary = true, and demotes the COVER video.
func TestIndex_GooglePixel_Sequential_VideoBoost_LateMain(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	folder := "pixel_sequential_videoboost_latemain"
	cfg := newGooglePixelStackConfig(t, folder)
	dir := filepath.Join(cfg.OriginalsPath(), folder)

	videoSource := filepath.Join(cfg.SamplesPath(), "blue-go-video.mp4")
	coverVideo := filepath.Join(dir, "PXL_20240930_180000123.VB-01.COVER.mp4")
	mainVideo := filepath.Join(dir, "PXL_20240930_180000123.VB-03.MAIN.mp4")

	writeGooglePixelVideo(t, videoSource, coverVideo, "VB-01.COVER")

	// Step 1: Index draft preview COVER video first.
	indexGooglePixelFolder(cfg, folder, false)

	var photos []entity.Photo
	require.NoError(t, entity.UnscopedDb().Where("photo_path = ?", folder).Find(&photos).Error)
	require.Len(t, photos, 1, "initial video entity for draft COVER video must exist")
	initialPhoto := photos[0]

	// Step 2: Add boosted MAIN video and re-index.
	writeGooglePixelVideo(t, videoSource, mainVideo, "VB-03.MAIN")
	indexGooglePixelFolder(cfg, folder, false)

	var photosAfter []entity.Photo
	require.NoError(t, entity.UnscopedDb().Where("photo_path = ? AND deleted_at IS NULL", folder).Find(&photosAfter).Error)
	assert.Len(t, photosAfter, 1, "should still have only 1 active photo after late boosted MAIN video arrives")

	var files []entity.File
	require.NoError(t, entity.UnscopedDb().Where("photo_id = ? AND file_root = ?", initialPhoto.ID, entity.RootOriginals).Find(&files).Error)
	assert.GreaterOrEqual(t, len(files), 2, "cover and main video files should be linked to initial photo")

	var coverFile, mainFile *entity.File
	for i := range files {
		if strings.HasSuffix(files[i].FileName, "VB-01.COVER.mp4") {
			coverFile = &files[i]
		} else if strings.HasSuffix(files[i].FileName, "VB-03.MAIN.mp4") {
			mainFile = &files[i]
		}
	}

	require.NotNil(t, coverFile, "VB-01.COVER video should be linked to photo")
	require.NotNil(t, mainFile, "VB-03.MAIN video should be linked to photo")

	assert.False(t, mainFile.FilePrimary, "video original file is never FilePrimary")
	assert.False(t, coverFile.FilePrimary, "draft COVER preview video must not be FilePrimary")

	previews := googlePixelStackPreviews(t, folder)
	require.Contains(t, previews, "PXL_20240930_180000123.VB-03.MAIN.mp4.jpg")
	require.Contains(t, previews, "PXL_20240930_180000123.VB-01.COVER.mp4.jpg")
	assert.True(t, previews["PXL_20240930_180000123.VB-03.MAIN.mp4.jpg"].FilePrimary, "late boosted MAIN video preview must be promoted to FilePrimary")
	assert.False(t, previews["PXL_20240930_180000123.VB-01.COVER.mp4.jpg"].FilePrimary, "draft COVER preview must not be FilePrimary")
}

// TestImport_GooglePixelCapture verifies that importing a Google Pixel Camera capture moves and stacks
// the files into a single photo entity with the primary JPEG marked as FilePrimary.
func TestImport_GooglePixelCapture(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	folder := "pixelimport"
	cfg := newGooglePixelStackConfig(t, folder)
	importDir := filepath.Join(cfg.ImportPath(), folder)

	coverName := "PXL_20240926_143000123.RAW-01.jpg"
	rawName := "PXL_20240926_143000123.RAW-02.ORIGINAL.dng"

	writeGooglePixelZeroGpsJpeg(t, filepath.Join(importDir, coverName))
	writeGooglePixelZeroGpsDng(t, cfg, filepath.Join(importDir, rawName))

	convert := NewConvert(cfg)
	NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert).Start(ImportOptionsMove(importDir, folder))

	stored := func(originalName string) (result entity.File) {
		t.Helper()
		require.NoError(t, entity.UnscopedDb().First(&result, "file_root = ? AND original_name = ?", entity.RootOriginals, originalName).Error)
		return result
	}

	coverFile := stored(coverName)
	rawFile := stored(rawName)

	assert.Equal(t, coverFile.PhotoID, rawFile.PhotoID, "both files must belong to the same photo")
	assert.True(t, coverFile.FilePrimary, "imported cover JPEG must be FilePrimary")
	assert.False(t, rawFile.FilePrimary, "imported companion RAW must not be FilePrimary")
}
