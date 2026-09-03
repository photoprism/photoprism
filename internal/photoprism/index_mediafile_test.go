package photoprism

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/thumb"
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

	// The package-wide test config points all test configs at one shared database,
	// so this test needs one of its own for reliable row counts.
	useTestDb(t, "index-dup-race")

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

	// The package-wide test config points all test configs at one shared database;
	// the database and storage must be isolated so the flash.jpg content does not
	// collide by hash with a row another test indexed with an explicit original name.
	useTestDb(t, "index-original-name")

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

// TestIndex_MediaFile_DualFisheye verifies that an Insta360 .insp original is recognized and
// recorded with the dual-fisheye projection, which also flags the photo as a panorama.
func TestIndex_MediaFile_DualFisheye(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	useTestDb(t, "index-dual-fisheye")

	cfg := config.NewMinimalTestConfigWithDb("index-dual-fisheye", filepath.Join(t.TempDir(), "storage"))

	oldCfg := Config()
	SetConfig(cfg)

	t.Cleanup(func() {
		SetConfig(oldCfg)
		oldCfg.RegisterDb()
	})

	convert := NewConvert(cfg)
	ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())
	opt := IndexOptionsSingle(cfg)
	opt.Convert = false // exercise original detection without the ffmpeg dewarp

	srcFile, err := NewMediaFile("testdata/insta360.insp")
	require.NoError(t, err)

	dst := filepath.Join(cfg.OriginalsPath(), "insta360.insp")
	require.NoError(t, srcFile.Copy(dst, false))

	mf, err := NewMediaFile(dst)
	require.NoError(t, err)
	hash := mf.Hash()

	// The full index flow creates the ExifTool JSON before indexing; it also supplies the JPEG
	// dimensions the panorama flag depends on.
	require.NoError(t, mf.CreateExifToolJson(convert))

	res := ind.MediaFile(mf, opt, "", "")
	require.True(t, res.Success())

	var file entity.File
	require.NoError(t, entity.UnscopedDb().First(&file, "file_hash = ?", hash).Error)
	assert.Equal(t, "dual-fisheye", file.FileProjection)

	var photo entity.Photo
	require.NoError(t, entity.UnscopedDb().First(&photo, "id = ?", file.PhotoID).Error)
	assert.True(t, photo.PhotoPanorama)
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

		useTestDb(t, "import-face-tags")
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

	useTestDb(t, "faces-only-recount")
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

// indexArchivedPhotoFixture reproduces the database state of issue #5766: a
// converted picture whose original (e.g. a .raw or .heic, still present in
// originals/) is stacked with a generated preview JPEG (the kind of file the
// converter places in storage/sidecar/), both belonging to one photo.
//
// The generated preview is indexed first so it becomes the photo's primary
// file; the original is indexed second as a stacked file, which links it to
// the same photo without replacing the primary.
//
// The test environment uses libvips stubs that cannot decode images, so a
// thumbnail is pre-seeded into the cache for every size the indexer would
// generate. This turns GenerateThumbnails into a no-op for the rest of the
// test and lets it exercise the archive/restore code path without image
// processing.
func indexArchivedPhotoFixture(t *testing.T, db, storageName, origBase string) (mfPreview *MediaFile, ind *Index, preview, original *entity.File, photo *entity.Photo) {
	t.Helper()

	useTestDb(t, db)

	cfg := config.NewMinimalTestConfigWithDb(db, filepath.Join(t.TempDir(), storageName))

	// MediaFile.Root() and the ExifTool cache resolve against the package-level
	// config, so it must point to the test config for this run.
	oldCfg := Config()
	SetConfig(cfg)

	t.Cleanup(func() {
		SetConfig(oldCfg)
		oldCfg.RegisterDb()
	})

	ind = NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())

	srcFile, err := NewMediaFile("testdata/flash.jpg")
	require.NoError(t, err)

	// The original picture stays in originals/ while the generated preview is
	// indexed as its stacked file (photo_stack > -1) and becomes the photo's
	// primary file.
	origName := filepath.Join(cfg.OriginalsPath(), storageName, origBase)
	require.NoError(t, srcFile.Copy(origName, false))

	// A converted preview is never byte-identical to its original, and the
	// indexer treats equal hashes as duplicates. Keep the hashes distinct.
	origBytes, err := os.ReadFile(origName)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(origName, append(origBytes, 0x00), fs.ModeFile))

	genName := filepath.Join(cfg.OriginalsPath(), storageName, origBase+".jpg")
	require.NoError(t, srcFile.Copy(genName, false))

	// Pre-seed the thumbnail cache so indexing runs without a working image
	// processor; otherwise GenerateThumbnails fails before the index reaches
	// the code path under test. The seed must be a decodable PNG because the
	// indexer reads back the cached Colors thumbnail, and a decode error
	// would drop the primary flag.
	var seedPng []byte
	{
		buf := &bytes.Buffer{}
		require.NoError(t, png.Encode(buf, image.NewRGBA(image.Rect(0, 0, 1, 1))))
		seedPng = buf.Bytes()
	}

	for _, name := range []string{origName, genName} {
		mf, err := NewMediaFile(name)
		require.NoError(t, err)
		thumbHash := mf.Hash()

		for _, sizeName := range thumb.Names {
			if size := thumb.Sizes[sizeName]; !size.Uncached() {
				fileName, err := size.FileName(thumbHash, cfg.ThumbCachePath())
				require.NoError(t, err)
				require.NoError(t, fs.MkdirAll(filepath.Dir(fileName)))
				require.NoError(t, os.WriteFile(fileName, seedPng, fs.ModeFile))
			}
		}
	}

	// The pre-seeded 1x1 color thumbnail is a valid image, so the color
	// detector would classify it with the TensorFlow stubs and panic on this
	// platform. Classification is not the code path under test.
	opt := IndexOptionsSingle(cfg)
	opt.GenerateLabels = false
	opt.DetectNsfw = false

	// Index the generated preview first so it becomes the primary file.
	mfPreview, err = NewMediaFile(genName)
	require.NoError(t, err)
	require.NoError(t, mfPreview.CreateExifToolJson(NewConvert(cfg)))
	res := ind.MediaFile(mfPreview, opt, "", "")
	require.True(t, res.Success(), "initial index of the preview must succeed: %v", res.Err)

	// Index the original second, stacked with the preview.
	mfOrig, err := NewMediaFile(origName)
	require.NoError(t, err)
	require.NoError(t, mfOrig.CreateExifToolJson(NewConvert(cfg)))
	res = ind.MediaFile(mfOrig, opt, "", "")
	require.True(t, res.Success(), "initial index of the original must succeed: %v", res.Err)

	var previewFile entity.File
	require.NoError(t, entity.UnscopedDb().First(&previewFile, "file_hash = ? AND file_name = ?", mfPreview.Hash(), filepath.Join(storageName, origBase+".jpg")).Error)
	require.True(t, previewFile.FilePrimary, "the generated preview must be the photo's primary file")

	var originalFile entity.File
	require.NoError(t, entity.UnscopedDb().First(&originalFile, "file_hash = ? AND file_name = ?", mfOrig.Hash(), filepath.Join(storageName, origBase)).Error)

	var photoRow entity.Photo
	require.NoError(t, entity.UnscopedDb().First(&photoRow, "id = ?", previewFile.PhotoID).Error)
	require.False(t, photoRow.AllFilesMissing(), "fixture must start with present files")

	return mfPreview, ind, &previewFile, &originalFile, &photoRow
}

// TestIndex_ArchivedPhotoKeepsArchiveState reproduces issue #5766: a photo the
// user archived must keep its deleted_at when a later index pass sees
// photo_quality == -1 (set by query.FlagHiddenPhotos) after a purge run
// marked the missing generated preview as absent. The original is still
// present, which separates a user-archived photo from one that was purged
// automatically.
func TestIndex_ArchivedPhotoKeepsArchiveState(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	_, ind, _, _, photo := indexArchivedPhotoFixture(t, "index-keep-archive", "keep-archive", "photo.raw")

	// The user archived the photo: Photo.Archive sets deleted_at and leaves
	// photo_quality untouched.
	require.NoError(t, photo.Archive())

	// A purge run then marked the missing generated preview as absent and
	// query.FlagHiddenPhotos set photo_quality to -1 because the primary file
	// is gone. The original is still present on disk.
	require.NoError(t, previewPurge(t, "keep-archive", "photo.raw"))
	require.NoError(t, photo.Update("photo_quality", -1))
	require.NoError(t, entity.UnscopedDb().First(photo, "id = ?", photo.ID).Error)

	require.NotNil(t, photo.DeletedAt, "fixture must start with an archived photo")
	require.Equal(t, -1, photo.PhotoQuality, "fixture must start with a hidden quality")
	require.False(t, photo.AllFilesMissing(), "fixture must keep a present file (the original)")

	// Re-index the present original. Its file row is found by path, the
	// primary preview is flagged missing, and photo_quality is -1.
	mfOrig, err := NewMediaFile(filepath.Join(ind.conf.OriginalsPath(), "keep-archive", "photo.raw"))
	require.NoError(t, err)
	opt := IndexOptionsSingle(Config())
	opt.GenerateLabels = false
	opt.DetectNsfw = false
	res := ind.MediaFile(mfOrig, opt, "", "")
	require.True(t, res.Success(), "re-index must succeed: %v", res.Err)

	var reloaded entity.Photo
	require.NoError(t, entity.UnscopedDb().First(&reloaded, "id = ?", photo.ID).Error)
	require.NotNil(t, reloaded.DeletedAt, "indexing must keep deleted_at for a user-archived photo whose quality is -1 while a file is still present")
}

// TestIndex_ArchivedHeicKeepsArchiveState covers the HEIC variant of issue
// #5766: a converted HEIC picture keeps its original in originals/ while the
// converter output is the JPEG sidecar, so the missing-file scenario applies
// to HEIC pictures as well and the archive state must be preserved.
func TestIndex_ArchivedHeicKeepsArchiveState(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	_, ind, _, _, photo := indexArchivedPhotoFixture(t, "index-keep-archive-heic", "keep-archive-heic", "photo.heic")

	require.NoError(t, photo.Archive())
	require.NoError(t, previewPurge(t, "keep-archive-heic", "photo.heic"))
	require.NoError(t, photo.Update("photo_quality", -1))
	require.NoError(t, entity.UnscopedDb().First(photo, "id = ?", photo.ID).Error)

	require.NotNil(t, photo.DeletedAt, "fixture must start with an archived photo")
	require.Equal(t, -1, photo.PhotoQuality, "fixture must start with a hidden quality")
	require.False(t, photo.AllFilesMissing(), "fixture must keep a present file (the original)")

	mfOrig, err := NewMediaFile(filepath.Join(ind.conf.OriginalsPath(), "keep-archive-heic", "photo.heic"))
	require.NoError(t, err)
	opt := IndexOptionsSingle(Config())
	opt.GenerateLabels = false
	opt.DetectNsfw = false
	res := ind.MediaFile(mfOrig, opt, "", "")
	require.True(t, res.Success(), "re-index must succeed: %v", res.Err)

	var reloaded entity.Photo
	require.NoError(t, entity.UnscopedDb().First(&reloaded, "id = ?", photo.ID).Error)
	require.NotNil(t, reloaded.DeletedAt, "indexing must keep deleted_at for a user-archived HEIC photo whose quality is -1 while a file is still present")
}

// previewPurge marks the generated preview file (the one that is not on disk
// anymore) as missing and non-primary, as the purge step does for files that
// disappeared from storage/sidecar/.
func previewPurge(t *testing.T, storageName, origBase string) (err error) {
	t.Helper()

	var preview entity.File

	if err = entity.UnscopedDb().First(&preview, "file_name = ?", filepath.Join(storageName, origBase+".jpg")).Error; err != nil {
		return err
	}

	return preview.Purge()
}

// TestIndex_AutoPurgedPhotoIsRestored guards the existing purpose of the
// restore branch: a photo purged automatically (all of its files missing)
// must be restored when one of its files is found again.
func TestIndex_AutoPurgedPhotoIsRestored(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	mfPreview, ind, preview, original, photo := indexArchivedPhotoFixture(t, "index-restore-purged", "restore-purged", "photo.raw")

	// Simulate the end state of an automatic purge: the photo carries
	// deleted_at and photo_quality -1 (Photo.Delete's effect) and every file
	// row is flagged missing (the purge step's effect on files that
	// disappeared from disk).
	require.NoError(t, photo.Archive())
	require.NoError(t, preview.Purge())
	require.NoError(t, original.Purge())
	require.NoError(t, photo.Update("photo_quality", -1))
	require.NoError(t, entity.UnscopedDb().First(photo, "id = ?", photo.ID).Error)

	require.NotNil(t, photo.DeletedAt, "fixture must start with a purged photo")
	require.Equal(t, -1, photo.PhotoQuality, "fixture must start with a hidden quality")
	require.True(t, photo.AllFilesMissing(), "fixture must start with all files missing")

	// Re-index the preview file that is present again on disk.
	opt := IndexOptionsSingle(Config())
	opt.GenerateLabels = false
	opt.DetectNsfw = false
	res := ind.MediaFile(mfPreview, opt, "", "")
	require.True(t, res.Success(), "re-index must succeed: %v", res.Err)

	var reloaded entity.Photo
	require.NoError(t, entity.UnscopedDb().First(&reloaded, "id = ?", photo.ID).Error)
	require.Nil(t, reloaded.DeletedAt, "indexing must restore a photo that was purged automatically")
}
