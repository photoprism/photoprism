package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestImportWorker_OriginalFileNames(t *testing.T) {
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
	t.Setenv("PHOTOPRISM_TEST_DRIVER", dsn.DriverSQLite3)
	t.Setenv("PHOTOPRISM_TEST_DSN", filepath.Join(t.TempDir(), "import-stacked-vectors.db"))

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
