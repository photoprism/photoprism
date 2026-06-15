package photoprism

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestIndex_MediaFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("FlashJpg", func(t *testing.T) {
		cfg := config.TestConfig()

		initErr := cfg.InitializeTestData()
		assert.NoError(t, initErr)

		convert := NewConvert(cfg)

		ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())
		indexOpt := IndexOptionsAll(cfg)
		mediaFile, err := NewMediaFile("testdata/flash.jpg")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "", mediaFile.metaData.Keywords.String())

		result := ind.MediaFile(mediaFile, indexOpt, "flash.jpg", "")

		words := mediaFile.metaData.Keywords.String()

		t.Logf("size in megapixel: %d", mediaFile.Megapixels())

		if _, limitErr := mediaFile.ExceedsResolution(cfg.ResolutionLimit()); limitErr != nil {
			t.Logf("index: %s", limitErr)
		}

		assert.Contains(t, words, "marienkäfer")
		assert.Contains(t, words, "burst")
		assert.Contains(t, words, "flash")
		assert.Contains(t, words, "panorama")
		assert.Equal(t, "Animal with green eyes on table burst", mediaFile.metaData.Caption)
		assert.Equal(t, IndexStatus("added"), result.Status)
	})
	t.Run("BlueGoVideoMp4", func(t *testing.T) {
		cfg := config.TestConfig()

		initErr := cfg.InitializeTestData()
		assert.NoError(t, initErr)

		convert := NewConvert(cfg)

		ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())
		indexOpt := IndexOptionsAll(cfg)
		mediaFile, err := NewMediaFile(cfg.SamplesPath() + "/blue-go-video.mp4")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "", mediaFile.metaData.Title)

		result := ind.UserMediaFile(mediaFile, indexOpt, "blue-go-video.mp4", "", entity.Admin.GetUID())

		assert.Equal(t, "Blue Gopher", mediaFile.metaData.Title)
		assert.Equal(t, IndexStatus("added"), result.Status)
	})

	t.Run("twoFiles", func(t *testing.T) {
		cfg := config.TestConfig()

		cfg.InitializeTestData()

		// Cleanup before we run as maybe the files have already been loaded.
		var err error
		for ok := true; ok; ok = (err == nil) {
			prephoto := entity.Photo{}
			err = entity.UnscopedSearchFirstPhoto(&prephoto, "original_name = ? OR (photo_lat = ? AND photo_lng = ?)", "beach_sand", -29.28247777777778, 31.44363611111111).Error
			if err == nil {
				DeletePhoto(&prephoto, true, true)
			}
		}

		convert := NewConvert(cfg)

		ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())
		indexOpt := IndexOptionsAll(cfg)
		mediaFile, err := NewMediaFile(cfg.SamplesPath() + "/beach_sand.jpg")
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "", mediaFile.metaData.Title)
		assert.Equal(t, "", mediaFile.metaData.CameraMake)

		result := ind.UserMediaFile(mediaFile, indexOpt, "beach_sand.jpg", "", entity.Admin.GetUID())

		assert.Equal(t, "", mediaFile.metaData.Title)
		assert.Equal(t, "Apple", mediaFile.metaData.CameraMake)

		photo := entity.Photo{}
		entity.Db().Model(&entity.Photo{}).Preload("Details").Where("original_name = 'beach_sand'").First(&photo)
		assert.Equal(t, "beach_sand", photo.OriginalName)
		quality := photo.PhotoQuality
		cameraid := photo.CameraID
		placeid := photo.PlaceID
		assert.Contains(t, photo.Details.Keywords, "beach")
		assert.Contains(t, photo.Details.Keywords, "sand")
		assert.Contains(t, photo.Details.Keywords, "blue")
		assert.Equal(t, IndexStatus("added"), result.Status)

		mediaFile, err = NewMediaFile(cfg.SamplesPath() + "/beach_sand.json")
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "", mediaFile.metaData.CameraMake)
		assert.Equal(t, "", mediaFile.metaData.Title)

		result = ind.UserMediaFile(mediaFile, indexOpt, "beach_sand.jpg", "", entity.Admin.GetUID())

		// This isn't a Primary file, so these should NOT be updated.
		assert.Equal(t, "", mediaFile.metaData.CameraMake)
		assert.Equal(t, "", mediaFile.metaData.Title)

		photo = entity.Photo{}
		entity.Db().Model(&entity.Photo{}).Preload("Details").Where("original_name = 'beach_sand'").First(&photo)
		assert.Equal(t, "beach_sand", photo.OriginalName)
		assert.Contains(t, photo.Details.Keywords, "beach")
		assert.Contains(t, photo.Details.Keywords, "sand")
		assert.Contains(t, photo.Details.Keywords, "blue")
		// Make sure that reading in a json file with the same details as the photo hasn't changed the data.
		assert.Equal(t, quality, photo.PhotoQuality)
		assert.Equal(t, cameraid, photo.CameraID)
		assert.Equal(t, placeid, photo.PlaceID)

		assert.Equal(t, IndexStatus("added"), result.Status)
	})

	t.Run("MediaRestoring and YAML", func(t *testing.T) {
		cfg := config.TestConfig()

		cfg.InitializeTestData()

		if fileNameResolved, err := fs.Resolve("testdata/sidecar/photoprism.yml"); err != nil {
			t.Fatal(err)
		} else {
			target := cfg.SidecarPath() + "/mediarestoring.yml"
			fs.Copy(fileNameResolved, target, true)
		}

		if fileNameResolved, err := fs.Resolve("testdata/photoprism.png"); err != nil {
			t.Fatal(err)
		} else {
			target := cfg.OriginalsPath() + "/mediarestoring.png"
			fs.Copy(fileNameResolved, target, true)
		}

		convert := NewConvert(cfg)

		photoUID := "psz10aeojfji0b86"
		photo := entity.NewPhoto(true)
		photo.PhotoUID = photoUID
		// Set the photo as MediaRestoring as it's not real.
		photo.PhotoType = entity.MediaRestoring
		// Set it as Purged
		photo.DeletedAt = gorm.DeletedAt{Valid: true, Time: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)}
		photo.PhotoQuality = -1
		if err := photo.Save(); err != nil {
			t.Fatal(err)
		}

		ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())
		indexOpt := IndexOptionsAll(cfg)
		mediaFile, err := NewMediaFile(cfg.OriginalsPath() + "/mediarestoring.png")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "", mediaFile.metaData.Keywords.String())

		result := ind.MediaFile(mediaFile, indexOpt, "mediarestoring.png", "")

		assert.Equal(t, IndexStatus("added"), result.Status)

		if found := entity.FindPhoto(entity.Photo{PhotoUID: photoUID}); found == nil {
			t.Fatal("Unable to find photo by UID")
		} else {
			assert.Equal(t, gorm.DeletedAt{}, found.DeletedAt)
			assert.NotEqual(t, -1, found.PhotoQuality)
			assert.Equal(t, "Elephant / South Africa / 2014", found.PhotoTitle)
			assert.NotEqual(t, entity.MediaRestoring, found.PhotoType)
		}
	})

	t.Run("Error", func(t *testing.T) {
		cfg := config.TestConfig()

		initErr := cfg.InitializeTestData()
		assert.NoError(t, initErr)

		convert := NewConvert(cfg)

		ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())
		indexOpt := IndexOptionsAll(cfg)

		result := ind.MediaFile(nil, indexOpt, "blue-go-video.mp4", "")
		assert.Equal(t, IndexStatus("failed"), result.Status)
	})
}

// TestIndex_UserMediaFile_ParallelDuplicates verifies that byte-identical files indexed
// concurrently by multiple workers result in exactly one photo and N-1 duplicate records.
func TestIndex_UserMediaFile_ParallelDuplicates(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// The package-wide PHOTOPRISM_TEST_DSN points all test configs at one shared
	// SQLite file, so the database must be isolated for reliable row counts.
	t.Setenv("PHOTOPRISM_TEST_DSN", filepath.Join(t.TempDir(), "index-dup-race.db"))

	cfg := config.NewMinimalTestConfigWithDb("index-dup-race", filepath.Join(t.TempDir(), "storage"))

	// MediaFile.Root() resolves paths against the package-level config, so it
	// must point to the test config for files to be detected as originals.
	oldCfg := Config()
	SetConfig(cfg)

	t.Cleanup(func() {
		SetConfig(oldCfg)
		oldCfg.RegisterDb()
	})

	testFile, err := NewMediaFile("testdata/flash.jpg")

	if err != nil {
		t.Fatal(err)
	}

	const numCopies = 3

	copyNames := make([]string, numCopies)

	for i := range copyNames {
		copyNames[i] = filepath.Join(cfg.OriginalsPath(), fmt.Sprintf("folder%d", i), "flash.jpg")

		if copyErr := testFile.Copy(copyNames[i], false); copyErr != nil {
			t.Fatal(copyErr)
		}
	}

	ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())
	indexOpt := IndexOptionsSingle(cfg)

	// The test database is seeded with entity fixtures, so all row counts are compared as deltas.
	var basePhotos, baseFiles, baseDuplicates int

	assert.NoError(t, entity.UnscopedDb().Model(&entity.Photo{}).Count(&basePhotos).Error)
	assert.NoError(t, entity.UnscopedDb().Model(&entity.File{}).Count(&baseFiles).Error)
	assert.NoError(t, entity.UnscopedDb().Model(&entity.Duplicate{}).Count(&baseDuplicates).Error)

	mediaFiles := make([]*MediaFile, numCopies)

	for i, name := range copyNames {
		if mediaFiles[i], err = NewMediaFile(name); err != nil {
			t.Fatal(err)
		}
	}

	results := make([]IndexResult, numCopies)
	start := make(chan struct{})

	var wg sync.WaitGroup

	for i := range mediaFiles {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			results[i] = ind.UserMediaFile(mediaFiles[i], indexOpt, "", "", entity.OwnerUnknown)
		}(i)
	}

	close(start)
	wg.Wait()

	added, duplicates := 0, 0

	for _, result := range results {
		switch result.Status {
		case IndexAdded:
			added++
		case IndexDuplicate:
			duplicates++
		default:
			t.Fatalf("unexpected index result %s (%v)", result.Status, result.Err)
		}
	}

	assert.Equal(t, 1, added)
	assert.Equal(t, numCopies-1, duplicates)

	var photoCount, fileCount, duplicateCount int

	assert.NoError(t, entity.UnscopedDb().Model(&entity.Photo{}).Count(&photoCount).Error)
	assert.NoError(t, entity.UnscopedDb().Model(&entity.File{}).Count(&fileCount).Error)
	assert.NoError(t, entity.UnscopedDb().Model(&entity.Duplicate{}).Count(&duplicateCount).Error)

	assert.Equal(t, basePhotos+1, photoCount)
	assert.Equal(t, baseFiles+1, fileCount)
	assert.Equal(t, baseDuplicates+numCopies-1, duplicateCount)
}

func TestIndexResult_Archived(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		r := &IndexResult{IndexArchived, nil, 5, "", 5, ""}
		assert.True(t, r.Archived())
	})
	t.Run("False", func(t *testing.T) {
		r := &IndexResult{IndexAdded, nil, 5, "", 5, ""}
		assert.False(t, r.Archived())
	})
}

func TestIndexResult_Skipped(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		r := &IndexResult{IndexSkipped, nil, 5, "", 5, ""}
		assert.True(t, r.Skipped())
	})
	t.Run("False", func(t *testing.T) {
		r := &IndexResult{IndexAdded, nil, 5, "", 5, ""}
		assert.False(t, r.Skipped())
	})
}
