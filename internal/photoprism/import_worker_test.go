package photoprism

import (
	"errors"
	"fmt"
	iofs "io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/meta"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestImportFailures verifies how failures of an import run are counted and reported.
func TestImportFailures(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		assert.NoError(t, (&ImportFailures{}).Err())
		var nilFailures *ImportFailures
		assert.NotPanics(t, func() { nilFailures.add(errors.New("failed")) })
		assert.NoError(t, nilFailures.Err())
	})
	t.Run("Incomplete", func(t *testing.T) {
		f := &ImportFailures{}
		f.add(errors.New("failed"))
		f.add(os.ErrPermission)
		assert.ErrorIs(t, f.Err(), ErrImportIncomplete)
		assert.Contains(t, f.Err().Error(), "(2)")
	})
	t.Run("NoSpace", func(t *testing.T) {
		f := &ImportFailures{}
		f.add(errors.New("failed"))
		f.add(&os.PathError{Op: "write", Path: "/originals/photo.jpg", Err: syscall.ENOSPC})
		assert.ErrorIs(t, f.Err(), status.ErrInsufficientStorage)
		assert.NotContains(t, f.Err().Error(), "/originals")
		quota := &ImportFailures{}
		quota.add(&os.PathError{Op: "write", Path: "photo.jpg", Err: syscall.EDQUOT})
		assert.ErrorIs(t, quota.Err(), status.ErrInsufficientStorage)
	})
	t.Run("NoSpaceText", func(t *testing.T) {
		f := &ImportFailures{}
		f.add(errors.New("no space left on device.jpg cannot be opened"))
		assert.ErrorIs(t, f.Err(), ErrImportIncomplete)
	})
}

// TestImportedContent verifies that only a regular file with the same content counts as imported.
func TestImportedContent(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "photo.jpg")
	require.NoError(t, os.WriteFile(file, []byte("content"), fs.ModeFile))
	hash := fs.Hash(file)
	link := filepath.Join(dir, "link.jpg")
	require.NoError(t, os.Symlink(file, link))
	other := filepath.Join(dir, "other.jpg")
	require.NoError(t, os.WriteFile(other, []byte("other"), fs.ModeFile))
	assert.True(t, importedContent(file, hash))
	assert.False(t, importedContent(file, ""))
	assert.False(t, importedContent(link, hash))
	assert.False(t, importedContent(other, hash))
	assert.False(t, importedContent(filepath.Join(dir, "missing.jpg"), hash))
	assert.False(t, importedContent(dir, hash))
}

func TestImportWorker_OriginalFileNames(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// Use the package-level config set in TestMain to avoid diverging
	// settings/paths from the code under test.
	cfg := Config()

	initErr := cfg.InitializeTestData()
	assert.NoError(t, initErr)

	convert := NewConvert(cfg)
	ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())
	imp := &Import{cfg, ind, convert, cfg.ImportAllow()}

	mediaFileName := cfg.SamplesPath() + "/beach_sand.jpg"
	mediaFile, err := NewMediaFile(mediaFileName)
	if err != nil {
		t.Fatal(err)
	}
	mediaFileName2 := cfg.SamplesPath() + "/beach_wood.jpg"
	mediaFile2, err2 := NewMediaFile(mediaFileName2)
	if err2 != nil {
		t.Fatal(err2)
	}
	mediaFileName3 := cfg.SamplesPath() + "/beach_colorfilter.jpg"
	mediaFile3, err3 := NewMediaFile(mediaFileName3)
	if err3 != nil {
		t.Fatal(err3)
	}
	relatedFiles := RelatedFiles{
		Files: MediaFiles{mediaFile, mediaFile2, mediaFile3},
		Main:  mediaFile,
	}

	jobs := make(chan ImportJob)
	done := make(chan bool)

	go func() {
		ImportWorker(jobs)
		done <- true
	}()

	jobs <- ImportJob{
		FileName:  mediaFile.FileName(),
		Related:   relatedFiles,
		IndexOpt:  IndexOptionsAll(cfg),
		ImportOpt: ImportOptionsCopy(cfg.ImportPath(), cfg.ImportDest()),
		Imp:       imp,
	}

	// Wait for job to finish.
	close(jobs)
	<-done

	var file entity.File
	res := entity.UnscopedDb().First(&file, "original_name = ?", mediaFileName)
	assert.Nil(t, res.Error)
	assert.Equal(t, mediaFileName, file.OriginalName)

	var file2 entity.File
	res = entity.UnscopedDb().First(&file2, "original_name = ?", mediaFileName2)
	assert.Nil(t, res.Error)
	assert.Equal(t, mediaFileName2, file2.OriginalName)

	var file3 entity.File
	res = entity.UnscopedDb().First(&file3, "original_name = ?", mediaFileName3)
	assert.Nil(t, res.Error)
	assert.Equal(t, mediaFileName3, file3.OriginalName)
}

// TestImportWorker_StackedVectorPreviews verifies that importing a stack with more than one
// convertible file (a base .svg plus a .touch.svg variant) generates a preview image for each
// of them, so the file that becomes the indexed primary always has a matching sidecar and the
// resulting photo is not hidden. Regression: the worker previously created only a single preview.
func TestImportWorker_StackedVectorPreviews(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// Isolate the database so the imported stack is the only photo present.
	useTestDb(t, "import-stacked-vectors")

	cfg := config.NewMinimalTestConfigWithDb("import-stacked-vectors", filepath.Join(t.TempDir(), "storage"))

	if !cfg.VectorEnabled() {
		t.Skip("requires vector support (rsvg-convert or ImageMagick)")
	}

	// MediaFile.Root() and RelatedFiles() resolve paths against the package-level config.
	oldCfg := Config()
	SetConfig(cfg)
	t.Cleanup(func() {
		SetConfig(oldCfg)
		oldCfg.RegisterDb()
	})

	// Reproduce the affected instance, which had stacking by file name enabled.
	cfg.Settings().Stack.Name = true

	// A base icon plus its full-bleed touch variant, made byte-unique (different fill) so they
	// are not treated as duplicates. The two files legitimately stack (shared name prefix).
	baseSvg := `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16"><rect width="16" height="16" fill="#010203"/></svg>`
	touchSvg := `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16"><rect width="16" height="16" fill="#040506"/></svg>`

	importDir := filepath.Join(t.TempDir(), "import")
	if err := os.MkdirAll(importDir, fs.ModeDir); err != nil {
		t.Fatal(err)
	}

	baseName := filepath.Join(importDir, "icon.svg")
	touchName := filepath.Join(importDir, "icon.touch.svg")
	if err := os.WriteFile(baseName, []byte(baseSvg), fs.ModeFile); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(touchName, []byte(touchSvg), fs.ModeFile); err != nil {
		t.Fatal(err)
	}

	// Build the related-file group exactly as the import walker does.
	baseFile, err := NewMediaFile(baseName)
	if err != nil {
		t.Fatal(err)
	}
	related, err := baseFile.RelatedFiles(cfg.Settings().StackSequences())
	if err != nil {
		t.Fatal(err)
	}
	assert.Len(t, related.Files, 2)

	convert := NewConvert(cfg)
	ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())
	imp := NewImport(cfg, ind, convert)

	jobs := make(chan ImportJob)
	done := make(chan bool)
	go func() {
		ImportWorker(jobs)
		done <- true
	}()
	jobs <- ImportJob{
		FileName:  baseName,
		Related:   related,
		IndexOpt:  IndexOptionsAll(cfg),
		ImportOpt: ImportOptionsMove(importDir, ""),
		Imp:       imp,
	}
	close(jobs)
	<-done

	// Both stacked SVGs are indexed, and each must have its own PNG preview sidecar so that
	// whichever file becomes the primary has a matching preview (there are no SVG fixtures,
	// so these are the only two vector files in the database).
	var svgFiles entity.Files
	if res := entity.UnscopedDb().Where("file_type = ?", string(fs.VectorSVG)).Find(&svgFiles); res.Error != nil {
		t.Fatal(res.Error)
	}
	assert.Len(t, svgFiles, 2)

	for _, file := range svgFiles {
		mf, mfErr := NewMediaFile(FileName(file.FileRoot, file.FileName))
		if mfErr != nil {
			t.Fatal(mfErr)
		}

		assert.True(t, mf.HasPreviewImage(), "each stacked SVG should have its own PNG preview (%s)", file.FileName)
		assert.Empty(t, file.FileError, "stacked SVG should index without a file error (%s)", file.FileName)
	}

	// The photo the stack maps to must be visible, not hidden (photo_quality >= 0).
	var photo entity.Photo
	if res := entity.UnscopedDb().First(&photo, "id = ?", svgFiles[0].PhotoID); res.Error != nil {
		t.Fatalf("photo not found: %s", res.Error)
	}

	assert.GreaterOrEqual(t, photo.PhotoQuality, 0, "stacked vector photo must not be hidden")
}

// TestImportWorker_StackedRawOrientation verifies that a stacked file converted on import gets its
// orientation from ExifTool when the native parser cannot read it.
func TestImportWorker_StackedRawOrientation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	useTestDb(t, "import-stacked-raw")

	cfg := config.NewMinimalTestConfigWithDb("import-stacked-raw", filepath.Join(t.TempDir(), "storage"))

	if !cfg.ExifToolEnabled() {
		t.Skip("ExifTool must be available for the RAW embedded-preview fallback")
	}

	oldCfg := Config()
	SetConfig(cfg)
	t.Cleanup(func() {
		SetConfig(oldCfg)
		oldCfg.RegisterDb()
	})

	// Embedded previews are the only RAW converter, and the native parser refuses every file.
	cfg.Options().DisableDarktable = true
	cfg.Options().DisableRawTherapee = true
	cfg.Options().DisableSips = true
	cfg.Settings().Stack.Name = true
	maxBytes := meta.ExifMaxFileBytes
	meta.ExifMaxFileBytes = 1024
	t.Cleanup(func() { meta.ExifMaxFileBytes = maxBytes })

	// Two byte-unique RAW copies tagged 8 that stack by their shared name prefix.
	importDir := filepath.Join(t.TempDir(), "import")
	require.NoError(t, os.MkdirAll(importDir, fs.ModeDir))
	mainName := filepath.Join(importDir, "shot.dng")
	altName := filepath.Join(importDir, "shot.alt.dng")

	for i, fileName := range []string{mainName, altName} {
		require.NoError(t, fs.Copy(filepath.Join(cfg.SamplesPath(), "canon_eos_6d.dng"), fileName, true))
		// #nosec G204 -- arguments are the configured ExifTool binary, fixed values, and a test file path.
		require.NoError(t, exec.Command(cfg.ExifToolBin(), "-q", "-overwrite_original", "-n", "-Orientation=8", fmt.Sprintf("-Artist=Copy %d", i), fileName).Run())
	}

	mainFile, err := NewMediaFile(mainName)
	require.NoError(t, err)
	related, err := mainFile.RelatedFiles(cfg.Settings().StackSequences())
	require.NoError(t, err)
	require.Len(t, related.Files, 2)

	for _, f := range related.Files {
		jsonName, jsonErr := f.ExifToolJsonName()
		require.NoError(t, jsonErr)
		require.NoError(t, os.RemoveAll(jsonName))
	}

	convert := NewConvert(cfg)
	ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())
	imp := NewImport(cfg, ind, convert)

	jobs := make(chan ImportJob)
	done := make(chan bool)
	go func() {
		ImportWorker(jobs)
		done <- true
	}()
	jobs <- ImportJob{
		FileName:  mainName,
		Related:   related,
		IndexOpt:  IndexOptionsAll(cfg),
		ImportOpt: ImportOptionsMove(importDir, ""),
		Imp:       imp,
	}
	close(jobs)
	<-done

	var rawFiles entity.Files
	require.NoError(t, entity.UnscopedDb().Where("original_name IN (?)", []string{"shot.dng", "shot.alt.dng"}).Find(&rawFiles).Error)
	require.Len(t, rawFiles, 2)

	for _, file := range rawFiles {
		rawName := FileName(file.FileRoot, file.FileName)
		jpegName := fs.ImageJpeg.FindFirst(rawName, []string{cfg.SidecarPath(), fs.PPHiddenPathname}, cfg.OriginalsPath(), false)
		require.NotEmpty(t, jpegName, "%s has no preview", file.FileName)
		assert.Equal(t, "8", exifOrientationTag(t, cfg, jpegName), "preview of %s", file.FileName)
	}
}

// TestImportWorker_TypeCheck verifies that files whose content does not match their extension are
// moved to the originals folder, so that no file of a stack is lost, but are not processed or indexed.
func TestImportWorker_TypeCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	jpg, err := os.ReadFile("testdata/2018-04-12 19_24_49.jpg")
	require.NoError(t, err)

	// importFiles writes the files to a new import folder and imports them as one group.
	importFiles := func(t *testing.T, name string, files map[string][]byte, main string) *config.Config {
		useTestDb(t, name)
		cfg := config.NewMinimalTestConfigWithDb(name, filepath.Join(t.TempDir(), "storage"))
		oldCfg := Config()
		SetConfig(cfg)
		t.Cleanup(func() {
			SetConfig(oldCfg)
			oldCfg.RegisterDb()
		})

		importDir := filepath.Join(t.TempDir(), "import")
		require.NoError(t, fs.MkdirAll(importDir))

		for fileName, data := range files {
			require.NoError(t, os.WriteFile(filepath.Join(importDir, fileName), data, fs.ModeFile)) //nolint:gosec // G703: test-owned path
		}

		mainFile, newErr := NewMediaFile(filepath.Join(importDir, main))
		require.NoError(t, newErr)
		related, relErr := mainFile.RelatedFiles(false)
		require.NoError(t, relErr)
		require.Len(t, related.Files, len(files))

		convert := NewConvert(cfg)
		ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())

		jobs := make(chan ImportJob)
		done := make(chan bool)
		go func() {
			ImportWorker(jobs)
			done <- true
		}()
		jobs <- ImportJob{
			FileName:  mainFile.FileName(),
			Related:   related,
			IndexOpt:  IndexOptionsAll(cfg),
			ImportOpt: ImportOptionsMove(importDir, ""),
			Imp:       NewImport(cfg, ind, convert),
		}
		close(jobs)
		<-done

		return cfg
	}

	// findFiles returns the names of the files with the given extension below the directory.
	findFiles := func(t *testing.T, dir, ext string) (names []string) {
		require.NoError(t, filepath.Walk(dir, func(fileName string, info os.FileInfo, walkErr error) error {
			if walkErr == nil && !info.IsDir() && strings.HasSuffix(fileName, ext) {
				names = append(names, fileName)
			}
			return nil
		}))
		return names
	}

	t.Run("RelatedFile", func(t *testing.T) {
		logger, ok := log.(*logrus.Logger)
		require.True(t, ok)
		hooks := logger.ReplaceHooks(make(logrus.LevelHooks))
		t.Cleanup(func() { logger.ReplaceHooks(hooks) })
		hook := test.NewLocal(logger)

		cfg := importFiles(t, "import-type-related", map[string][]byte{
			"photo.jpg":       jpg,
			"photo.edit.webp": append(append([]byte(nil), jpg...), []byte(t.Name())...),
		}, "photo.jpg")

		jpgFiles := findFiles(t, cfg.OriginalsPath(), ".jpg")
		require.Len(t, jpgFiles, 1)
		webpFiles := findFiles(t, cfg.OriginalsPath(), ".webp")
		require.Len(t, webpFiles, 1, "the related file is moved to originals")
		assert.Empty(t, findFiles(t, cfg.SidecarPath(), ".edit.webp.jpg"), "no preview is created for it")

		_, findErr := entity.FirstFileByHash(fs.Hash(jpgFiles[0]))
		assert.NoError(t, findErr, "the main file is indexed")
		_, findErr = entity.FirstFileByHash(fs.Hash(webpFiles[0]))
		assert.Error(t, findErr, "the related file is not indexed")

		// The related file is only moved and reported, so no converter or other tool reads it.
		destName := filepath.Base(webpFiles[0])
		reported := false
		for _, entry := range hook.AllEntries() {
			switch {
			case !strings.Contains(entry.Message, destName):
				continue
			case strings.Contains(entry.Message, "was not indexed"):
				reported = true
			default:
				assert.Contains(t, entry.Message, "moving related")
			}
		}
		assert.True(t, reported)
	})
	t.Run("RelatedFileOfVideo", func(t *testing.T) {
		// The preview of a video has its own name, so the related file is not covered by it.
		mov, readErr := os.ReadFile(filepath.Join(fs.Abs("../../assets/samples"), "earth.mov"))
		require.NoError(t, readErr)

		logger, ok := log.(*logrus.Logger)
		require.True(t, ok)
		hooks := logger.ReplaceHooks(make(logrus.LevelHooks))
		t.Cleanup(func() { logger.ReplaceHooks(hooks) })
		hook := test.NewLocal(logger)

		cfg := importFiles(t, "import-type-video", map[string][]byte{
			"clip.mov":       mov,
			"clip.edit.webp": append(append([]byte(nil), jpg...), []byte(t.Name())...),
		}, "clip.mov")

		webpFiles := findFiles(t, cfg.OriginalsPath(), ".webp")
		require.Len(t, webpFiles, 1, "the related file is moved to originals")
		_, findErr := entity.FirstFileByHash(fs.Hash(webpFiles[0]))
		assert.Error(t, findErr, "the related file is not indexed")

		destName := filepath.Base(webpFiles[0])
		for _, entry := range hook.AllEntries() {
			if strings.Contains(entry.Message, destName) && !strings.Contains(entry.Message, "was not indexed") {
				assert.Contains(t, entry.Message, "moving related")
			}
		}
	})
	t.Run("RelatedVideoMetadata", func(t *testing.T) {
		// Checking the type of a related video must not keep its ExifTool metadata from the index.
		heic, readErr := os.ReadFile(filepath.Join(fs.Abs("../../assets/samples"), "iphone_7.heic"))
		require.NoError(t, readErr)
		mp4, readErr := os.ReadFile(filepath.Join(fs.Abs("../../assets/samples"), "gopher-video.mp4"))
		require.NoError(t, readErr)

		cfg := importFiles(t, "import-type-video-meta", map[string][]byte{"live.heic": heic, "live.mp4": mp4}, "live.heic")
		mp4Files := findFiles(t, cfg.OriginalsPath(), ".mp4")
		require.Len(t, mp4Files, 1)

		file, findErr := entity.FirstFileByHash(fs.Hash(mp4Files[0]))
		require.NoError(t, findErr)
		assert.Equal(t, "avc1", file.FileCodec)
		assert.Positive(t, file.FileDuration)
	})
	t.Run("MainFile", func(t *testing.T) {
		cfg := importFiles(t, "import-type-main", map[string][]byte{
			"photo.webp": append(append([]byte(nil), jpg...), []byte(t.Name())...),
		}, "photo.webp")

		webpFiles := findFiles(t, cfg.OriginalsPath(), ".webp")
		require.Len(t, webpFiles, 1, "the main file is moved to originals")
		assert.Empty(t, findFiles(t, cfg.SidecarPath(), ".webp.jpg"), "no preview is created for it")
		assert.Empty(t, findFiles(t, cfg.CachePath(), ".json"), "no metadata is extracted from it")

		_, findErr := entity.FirstFileByHash(fs.Hash(webpFiles[0]))
		assert.Error(t, findErr, "the main file is not indexed")
	})
}

// TestImport_TypeCheck verifies that the import walk moves files whose content does not match their
// extension with their stack, without extracting their metadata.
func TestImport_TypeCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	jpg, err := os.ReadFile("testdata/2018-04-12 19_24_49.jpg")
	require.NoError(t, err)

	// startImport writes the files to a new import folder and imports it with Import.Start.
	startImport := func(t *testing.T, name string, files map[string][]byte) *config.Config {
		useTestDb(t, name)
		cfg := config.NewMinimalTestConfigWithDb(name, filepath.Join(t.TempDir(), "storage"))
		oldCfg := Config()
		SetConfig(cfg)
		t.Cleanup(func() {
			SetConfig(oldCfg)
			oldCfg.RegisterDb()
		})

		importDir := filepath.Join(cfg.ImportPath(), name)
		require.NoError(t, fs.MkdirAll(importDir))

		for fileName, data := range files {
			require.NoError(t, os.WriteFile(filepath.Join(importDir, fileName), data, fs.ModeFile)) //nolint:gosec // G703: test-owned path
		}

		convert := NewConvert(cfg)
		NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert).Start(ImportOptionsMove(importDir, ""))

		return cfg
	}

	// assertMovedNotRead requires a single moved file with the extension, without ExifTool metadata or index entry.
	assertMovedNotRead := func(t *testing.T, cfg *config.Config, ext string) {
		var moved []string
		require.NoError(t, filepath.Walk(cfg.OriginalsPath(), func(fileName string, info os.FileInfo, walkErr error) error {
			if walkErr == nil && !info.IsDir() && strings.HasSuffix(fileName, ext) {
				moved = append(moved, fileName)
			}
			return nil
		}))
		require.Len(t, moved, 1)
		hash := fs.Hash(moved[0])
		jsonName, jsonErr := ExifToolCacheName(hash)
		require.NoError(t, jsonErr)
		assert.NoFileExists(t, jsonName)
		_, findErr := entity.FirstFileByHash(hash)
		assert.Error(t, findErr)
	}

	t.Run("MainFile", func(t *testing.T) {
		cfg := startImport(t, "import-start-type-main", map[string][]byte{
			"photo.webp": append(append([]byte(nil), jpg...), []byte(t.Name())...),
		})
		assertMovedNotRead(t, cfg, ".webp")
	})
	t.Run("RelatedFileFirst", func(t *testing.T) {
		cfg := startImport(t, "import-start-type-related", map[string][]byte{
			"photo.edit.webp": append(append([]byte(nil), jpg...), []byte(t.Name())...),
			"photo.jpg":       jpg,
		})
		assertMovedNotRead(t, cfg, ".webp")
	})
}

// TestImportWorker_SidecarPreview verifies that a preview in the sidecar folder is not imported with a file from the
// import folder.
func TestImportWorker_SidecarPreview(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	useTestDb(t, "import-sidecar-preview")

	cfg := config.NewMinimalTestConfigWithDb("import-sidecar-preview", filepath.Join(t.TempDir(), "storage"))
	oldCfg := Config()
	SetConfig(cfg)
	t.Cleanup(func() {
		SetConfig(oldCfg)
		oldCfg.RegisterDb()
	})

	// A HEIC without a JPEG of its own, and an unrelated JPEG at the sidecar name of its absolute path.
	importDir := filepath.Join(cfg.ImportPath(), "trip")
	heicName := filepath.Join(importDir, "IMG_0001.heic")
	foreignName := filepath.Join(cfg.SidecarPath(), importDir, "IMG_0001.heic.jpg")
	require.NoError(t, os.MkdirAll(importDir, fs.ModeDir))
	require.NoError(t, os.MkdirAll(filepath.Dir(foreignName), fs.ModeDir))
	require.NoError(t, fs.Copy(filepath.Join(cfg.SamplesPath(), "iphone_7.heic"), heicName, false))
	require.NoError(t, fs.Copy(filepath.Join(cfg.SamplesPath(), "beach_sand.jpg"), foreignName, false))

	mainFile, err := NewMediaFile(heicName)
	require.NoError(t, err)
	related, err := mainFile.RelatedFiles(false)
	require.NoError(t, err)

	convert := NewConvert(cfg)
	imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)
	jobs := make(chan ImportJob, 1)
	jobs <- ImportJob{FileName: heicName, Related: related, IndexOpt: IndexOptionsAll(cfg), ImportOpt: ImportOptionsMove(cfg.ImportPath(), ""), Imp: imp}
	close(jobs)
	ImportWorker(jobs)

	// The unrelated preview stays where it is, no copy of it reaches originals, and the HEIC is imported.
	assert.FileExists(t, foreignName)
	foreignHash := fs.Hash(foreignName)
	heicImported := false
	err = filepath.WalkDir(cfg.OriginalsPath(), func(name string, d iofs.DirEntry, walkErr error) error {
		if walkErr == nil && !d.IsDir() {
			assert.NotEqual(t, foreignHash, fs.Hash(name), name)
			heicImported = heicImported || strings.EqualFold(filepath.Ext(name), ".heic")
		}
		return nil
	})
	require.NoError(t, err)
	assert.True(t, heicImported)

	var files entity.Files
	require.NoError(t, entity.UnscopedDb().Where("file_hash = ?", foreignHash).Find(&files).Error)
	assert.Empty(t, files)
}

// TestImportWorker_ArchivedBackup checks that the related files of a file restored as archived from a backup
// at its destination are not indexed as a photo of their own when archived photos are skipped.
func TestImportWorker_ArchivedBackup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	useTestDb(t, "import-archived-backup")

	cfg := config.NewMinimalTestConfigWithDb("import-archived-backup", filepath.Join(t.TempDir(), "storage"))
	oldCfg := Config()
	SetConfig(cfg)
	t.Cleanup(func() {
		SetConfig(oldCfg)
		oldCfg.RegisterDb()
	})

	importDir := filepath.Join(cfg.ImportPath(), "clips")
	clipName := filepath.Join(importDir, "clip.mp4")
	require.NoError(t, os.MkdirAll(importDir, fs.ModeDir))
	require.NoError(t, fs.Copy(filepath.Join(cfg.SamplesPath(), "example.mp4"), clipName, false))

	mainFile, err := NewMediaFile(clipName)
	require.NoError(t, err)
	related, err := mainFile.RelatedFiles(false)
	require.NoError(t, err)

	convert := NewConvert(cfg)
	imp := NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert)

	// Place a backup of an archived photo at the destination, as after restoring the originals only. The
	// destination depends on the capture time, which the import reads with ExifTool.
	require.NoError(t, mainFile.CreateExifToolJson(convert))
	destName, err := imp.DestinationFilename(mainFile, mainFile, "")
	require.NoError(t, err)
	destPath := filepath.Dir(fs.RelName(destName, cfg.OriginalsPath()))
	require.NoError(t, os.MkdirAll(filepath.Dir(destName), fs.ModeDir))
	photoUID := rnd.GenerateUID(entity.PhotoUID)
	require.NoError(t, os.WriteFile(fs.StripExt(destName)+fs.ExtYml, []byte("UID: "+photoUID+
		"\nType: video\nQuality: 3\nDeletedAt: 2022-10-18T08:15:21Z\n"), fs.ModeFile))

	jobs := make(chan ImportJob, 1)
	jobs <- ImportJob{FileName: clipName, Related: related, IndexOpt: IndexOptionsAll(cfg), ImportOpt: ImportOptionsMove(cfg.ImportPath(), ""), Imp: imp}
	close(jobs)
	ImportWorker(jobs)

	// The photo is restored from its backup, even though the import skips archived photos, and stays archived.
	var photos entity.Photos
	require.NoError(t, entity.UnscopedDb().Where("photo_path = ?", destPath).Find(&photos).Error)
	require.Len(t, photos, 1)
	assert.Equal(t, photoUID, photos[0].PhotoUID)
	assert.True(t, photos[0].IsArchived(), "photo stays archived")

	var files []string
	require.NoError(t, entity.UnscopedDb().Model(&entity.File{}).Where("photo_id = ? AND deleted_at IS NULL", photos[0].ID).Pluck("file_name", &files).Error)
	mainName := fs.RelName(destName, cfg.OriginalsPath())
	assert.ElementsMatch(t, []string{mainName, fs.StripExt(mainName) + fs.ExtYml, mainName + fs.ExtJpeg}, files)
}
