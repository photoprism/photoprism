package photoprism

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/dsn"
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
	// database, so it must be isolated for reliable row counts.
	t.Setenv("PHOTOPRISM_TEST_DRIVER", dsn.DriverSQLite3)
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

// TestIndex_IndexedFileOriginalName verifies that files placed directly in
// originals/ and indexed (never imported) are not assigned an OriginalName,
// so the displayed card name keeps following the current file name after a
// rename and re-index.
func TestIndex_IndexedFileOriginalName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// The package-wide PHOTOPRISM_TEST_DSN points all test configs at one
	// shared database; it and the storage must be isolated so the flash.jpg
	// content does not collide by hash with a row another test indexed with
	// an explicit original name.
	t.Setenv("PHOTOPRISM_TEST_DRIVER", dsn.DriverSQLite3)
	t.Setenv("PHOTOPRISM_TEST_DSN", filepath.Join(t.TempDir(), "index-original-name.db"))

	cfg := config.NewMinimalTestConfigWithDb("index-original-name", filepath.Join(t.TempDir(), "storage"))

	// MediaFile.Root() and the ExifTool cache resolve against the package-level
	// config, so it must point to the test config for this run.
	oldCfg := Config()
	SetConfig(cfg)

	t.Cleanup(func() {
		SetConfig(oldCfg)
		oldCfg.RegisterDb()
	})

	convert := NewConvert(cfg)
	ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())
	opt := IndexOptionsSingle(cfg)

	srcFile, err := NewMediaFile("testdata/flash.jpg")
	require.NoError(t, err)

	first := filepath.Join(cfg.OriginalsPath(), "indexed-original-name", "indexed-photo.jpg")
	require.NoError(t, srcFile.Copy(first, false))

	mf1, err := NewMediaFile(first)
	require.NoError(t, err)
	hash := mf1.Hash()

	// The ExifTool JSON cache is keyed by content hash and records the current
	// file name; index_main.go creates it before indexing the media file.
	require.NoError(t, mf1.CreateExifToolJson(convert))

	// Plain index: callers pass an empty originalName for indexed files.
	res1 := ind.MediaFile(mf1, opt, "", "")
	require.True(t, res1.Success())

	var file1 entity.File
	require.NoError(t, entity.UnscopedDb().First(&file1, "file_hash = ?", hash).Error)
	assert.Empty(t, file1.OriginalName, "freshly indexed file must not carry an OriginalName")

	// Rename the file in originals/. The content (and therefore the hash and
	// the cached ExifTool JSON, which still records the old name) is unchanged,
	// which is what previously leaked a stale name into OriginalName.
	renamed := filepath.Join(cfg.OriginalsPath(), "indexed-original-name", "renamed-photo.jpg")
	require.NoError(t, fs.Move(first, renamed, false))

	mf2, err := NewMediaFile(renamed)
	require.NoError(t, err)
	require.NoError(t, mf2.CreateExifToolJson(convert))

	res2 := ind.MediaFile(mf2, opt, "", "")
	require.True(t, res2.Success())

	var file2 entity.File
	require.NoError(t, entity.UnscopedDb().First(&file2, "file_hash = ?", hash).Error)
	assert.Equal(t, "indexed-original-name/renamed-photo.jpg", file2.FileName)
	assert.Empty(t, file2.OriginalName, "re-indexed renamed file must not pick up a stale OriginalName")
}

func TestIndex_MediaFile_ImportFaceTags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// indexSidecar indexes the "Cara" sidecar fixture in an isolated database and
	// storage with the given face flags, then returns the file's markers. The
	// isolation mirrors TestIndex_MediaFile_OriginalName so the fixture hash does
	// not collide with rows another test indexed.
	indexSidecar := func(t *testing.T, detectFaces, importFaceTags bool) entity.Markers {
		t.Helper()

		t.Setenv("PHOTOPRISM_TEST_DRIVER", dsn.DriverSQLite3)
		t.Setenv("PHOTOPRISM_TEST_DSN", filepath.Join(t.TempDir(), "import-face-tags.db"))
		cfg := config.NewMinimalTestConfigWithDb("import-face-tags", filepath.Join(t.TempDir(), "storage"))

		// collectXmpFaces resolves the sidecar via the package-level config.
		oldCfg := Config()
		SetConfig(cfg)
		t.Cleanup(func() {
			SetConfig(oldCfg)
			oldCfg.RegisterDb()
		})

		// Copy the JPEG and its .xmp sidecar into originals so collectXmpFaces
		// finds the "Cara" region next to the primary file.
		dstDir := filepath.Join(cfg.OriginalsPath(), "xmp-faces")
		jpg := filepath.Join(dstDir, "sidecar.jpg")
		src, err := NewMediaFile("testdata/xmp-faces/sidecar.jpg")
		require.NoError(t, err)
		require.NoError(t, src.Copy(jpg, false))
		require.NoError(t, fs.Copy("testdata/xmp-faces/sidecar.jpg.xmp", filepath.Join(dstDir, "sidecar.jpg.xmp"), false))

		ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())
		opt := IndexOptionsSingle(cfg)
		opt.DetectFaces = detectFaces
		opt.ImportFaceTags = importFaceTags

		mf, err := NewMediaFile(jpg)
		require.NoError(t, err)

		result := ind.MediaFile(mf, opt, "", "")
		require.True(t, result.Success(), "index must succeed: %v", result.Err)
		require.NotEmpty(t, result.FileUID)

		markers, err := entity.FindMarkers(result.FileUID)
		require.NoError(t, err)
		return markers
	}

	hasXmpName := func(markers entity.Markers, name string) bool {
		for i := range markers {
			if markers[i].MarkerSrc == entity.SrcXmp && markers[i].MarkerName == name {
				return true
			}
		}
		return false
	}

	t.Run("ImportsWhenDetectionDisabled", func(t *testing.T) {
		markers := indexSidecar(t, false, true)
		assert.True(t, hasXmpName(markers, "Cara"), "XMP face tag must import with AI detection off, got %+v", markers)
	})
	t.Run("SkipsWhenToggleOff", func(t *testing.T) {
		markers := indexSidecar(t, false, false)
		assert.False(t, hasXmpName(markers, "Cara"), "no XMP marker may be created when the import toggle is off")
	})
}

// TestIndex_MediaFile_FacesOnlyRecountsAfterDelete verifies that a faces-only
// re-index recomputes and persists the photo face count when a removed XMP
// region deletes its marker, even though the deletion leaves no unsaved marker.
func TestIndex_MediaFile_FacesOnlyRecountsAfterDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Setenv("PHOTOPRISM_TEST_DSN", filepath.Join(t.TempDir(), "faces-only-recount.db"))
	cfg := config.NewMinimalTestConfigWithDb("faces-only-recount", filepath.Join(t.TempDir(), "storage"))
	oldCfg := Config()
	SetConfig(cfg)
	t.Cleanup(func() {
		SetConfig(oldCfg)
		oldCfg.RegisterDb()
	})

	dstDir := filepath.Join(cfg.OriginalsPath(), "xmp-faces")
	jpg := filepath.Join(dstDir, "sidecar.jpg")
	xmp := filepath.Join(dstDir, "sidecar.jpg.xmp")
	src, err := NewMediaFile("testdata/xmp-faces/sidecar.jpg")
	require.NoError(t, err)
	require.NoError(t, src.Copy(jpg, false))
	require.NoError(t, fs.Copy("testdata/xmp-faces/sidecar.jpg.xmp", xmp, false))

	ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())

	// Phase 1: import the sidecar's "Cara" region and confirm the face count.
	opt := IndexOptionsSingle(cfg)
	opt.ImportFaceTags = true
	mf, err := NewMediaFile(jpg)
	require.NoError(t, err)
	result := ind.MediaFile(mf, opt, "", "")
	require.True(t, result.Success(), "initial index must succeed: %v", result.Err)
	require.NotEmpty(t, result.PhotoUID)

	markers, err := entity.FindMarkers(result.FileUID)
	require.NoError(t, err)
	require.Len(t, markers, 1, "the sidecar region must import once")

	var before entity.Photo
	require.NoError(t, entity.Db().Where("photo_uid = ?", result.PhotoUID).First(&before).Error)
	require.Equal(t, 1, before.PhotoFaces, "photo face count must reflect the imported marker")

	// Phase 2: empty the sidecar's region list and re-index faces-only. The
	// container still declares that regions are tracked, so the removed region
	// deletes its marker; the count must be recomputed and persisted to 0.
	require.NoError(t, os.WriteFile(xmp, []byte(emptyRegionListXmp), fs.ModeFile)) //nolint:gosec // isolated test path

	facesOpt := IndexOptionsFacesOnly(cfg)
	facesOpt.ImportFaceTags = true
	mf2, err := NewMediaFile(jpg)
	require.NoError(t, err)
	res2 := ind.MediaFile(mf2, facesOpt, "", "")
	require.NotEqual(t, IndexFailed, res2.Status, "faces-only re-index must not fail: %v", res2.Err)

	remaining, err := entity.FindMarkers(result.FileUID)
	require.NoError(t, err)
	assert.Empty(t, remaining, "the removed region's XMP marker must be deleted")

	var after entity.Photo
	require.NoError(t, entity.Db().Where("photo_uid = ?", result.PhotoUID).First(&after).Error)
	assert.Equal(t, 0, after.PhotoFaces, "photo face count must be recomputed after a delete-only faces run")
}
