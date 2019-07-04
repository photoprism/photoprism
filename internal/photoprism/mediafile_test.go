package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestMediaFile_RelatedFiles(t *testing.T) {
	conf := config.TestConfig()

	mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/canon_eos_6d.dng")

	assert.Nil(t, err)

	expectedBaseFilename := conf.ExamplesPath() + "/canon_eos_6d"

	related, _, err := mediaFile.RelatedFiles()

	assert.Nil(t, err)

	assert.Len(t, related, 3)

	for _, result := range related {
		t.Logf("Filename: %s", result.Filename())

		filename := result.Filename()

		extension := result.Extension()

		baseFilename := filename[0 : len(filename)-len(extension)]

		assert.Equal(t, expectedBaseFilename, baseFilename)
	}
}

func TestMediaFile_RelatedFiles_Ordering(t *testing.T) {
	conf := config.TestConfig()

	mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/IMG_4120.JPG")

	assert.Nil(t, err)

	related, _, err := mediaFile.RelatedFiles()

	assert.Nil(t, err)

	assert.Len(t, related, 5)

	assert.Equal(t, conf.ExamplesPath()+"/IMG_4120.AAE", related[0].Filename())
	assert.Equal(t, conf.ExamplesPath()+"/IMG_4120.JPG", related[1].Filename())

	for _, result := range related {
		filename := result.Filename()
		t.Logf("Filename: %s", filename)
	}
}

func TestMediaFile_EditedFilename(t *testing.T) {
	conf := config.TestConfig()

	t.Run("IMG_4120.JPG", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/IMG_4120.JPG"); err == nil {

			assert.Nil(t, err)
			assert.Equal(t, conf.ExamplesPath()+"/IMG_E4120.JPG", mediaFile.EditedFilename())
		} else {
			t.Error(err)
		}
	})

	t.Run("fern_green.jpg", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/fern_green.jpg"); err == nil {

			assert.Nil(t, err)
			assert.Equal(t, "", mediaFile.EditedFilename())
		} else {
			t.Error(err)
		}
	})
}

func TestMediaFile_MimeType(t *testing.T) {
	conf := config.TestConfig()

	t.Run("elephants.jpg", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg"); err == nil {

			assert.Nil(t, err)
			assert.Equal(t, "image/jpeg", mediaFile.MimeType())
		} else {
			t.Error(err)
		}
	})

	t.Run("canon_eos_6d.dng", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/canon_eos_6d.dng"); err == nil {

			assert.Nil(t, err)
			assert.Equal(t, "application/octet-stream", mediaFile.MimeType())
		} else {
			t.Error(err)
		}
	})

	t.Run("iphone_7.xmp", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.xmp"); err == nil {

			assert.Nil(t, err)
			assert.Equal(t, "text/plain; charset=utf-8", mediaFile.MimeType())
		} else {
			t.Error(err)
		}
	})

	t.Run("iphone_7.json", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.json"); err == nil {

			assert.Nil(t, err)
			assert.Equal(t, "text/plain; charset=utf-8", mediaFile.MimeType())
		} else {
			t.Error(err)
		}
	})

	t.Run("iphone_7.heic", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic"); err == nil {

			assert.Nil(t, err)
			assert.Equal(t, "application/octet-stream", mediaFile.MimeType())
		} else {
			t.Error(err)
		}
	})

	t.Run("IMG_4120.AAE", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/IMG_4120.AAE"); err == nil {

			assert.Nil(t, err)
			assert.Equal(t, "text/xml; charset=utf-8", mediaFile.MimeType())
		} else {
			t.Error(err)
		}
	})
}

func TestMediaFile_Exists(t *testing.T) {
	conf := config.TestConfig()

	mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/cat_black.jpg")
	assert.Nil(t, err)
	assert.NotNil(t, mediaFile)
	assert.True(t, mediaFile.Exists())

	mediaFile, err = NewMediaFile(conf.ExamplesPath() + "/xxz.jpg")
	assert.NotNil(t, err)
	assert.Nil(t, mediaFile)
}
