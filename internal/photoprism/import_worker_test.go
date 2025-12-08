package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
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

	mediaFileName := cfg.ExamplesPath() + "/beach_sand.jpg"
	mediaFile, err := NewMediaFile(mediaFileName)
	if err != nil {
		t.Fatal(err)
	}
	mediaFileName2 := cfg.ExamplesPath() + "/beach_wood.jpg"
	mediaFile2, err2 := NewMediaFile(mediaFileName2)
	if err2 != nil {
		t.Fatal(err2)
	}
	mediaFileName3 := cfg.ExamplesPath() + "/beach_colorfilter.jpg"
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
