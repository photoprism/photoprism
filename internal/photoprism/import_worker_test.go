package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestImportWorker_OriginalFileNames(t *testing.T) {
	conf := config.TestConfig()

	if err := conf.InitializeTestData(); err != nil {
		t.Fatal(err)
	}

	tf := classify.New(conf.AssetsPath(), conf.DisableTensorFlow())
	nd := nsfw.New(conf.NSFWModelPath())
	fn := face.NewNet(conf.FaceNetModelPath(), "", conf.DisableTensorFlow())
	convert := NewConvert(conf)
	ind := NewIndex(conf, tf, nd, fn, convert, NewFiles(), NewPhotos())
	imp := &Import{conf, ind, convert}

	mediaFileName := conf.ExamplesPath() + "/beach_sand.jpg"
	mediaFile, err := NewMediaFile(mediaFileName)
	if err != nil {
		t.Fatal(err)
	}
	mediaFileName2 := conf.ExamplesPath() + "/beach_wood.jpg"
	mediaFile2, err2 := NewMediaFile(mediaFileName2)
	if err2 != nil {
		t.Fatal(err2)
	}
	mediaFileName3 := conf.ExamplesPath() + "/beach_colorfilter.jpg"
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
		IndexOpt:  IndexOptionsAll(),
		ImportOpt: ImportOptionsCopy(conf.ImportPath(), conf.ImportDest()),
		Imp:       imp,
	}

	// Wait for job to finish.
	close(jobs)
	<-done

	var file entity.File
	res := entity.UnscopedDb().First(&file, "original_name = ?", mediaFileName)
	assert.Nil(t, res.Error)
	assert.Equal(t, file.OriginalName, mediaFileName)

	var file2 entity.File
	res = entity.UnscopedDb().First(&file2, "original_name = ?", mediaFileName2)
	assert.Nil(t, res.Error)
	assert.Equal(t, file2.OriginalName, mediaFileName2)

	var file3 entity.File
	res = entity.UnscopedDb().First(&file3, "original_name = ?", mediaFileName3)
	assert.Nil(t, res.Error)
	assert.Equal(t, file3.OriginalName, mediaFileName3)
}
