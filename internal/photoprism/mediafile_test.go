package photoprism

import (
	"github.com/photoprism/photoprism/internal/util"
	"os"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestMediaFile_RelatedFiles(t *testing.T) {
	conf := config.TestConfig()

	t.Run("canon_eos_6d.dng", func(t *testing.T) {
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
	})

	t.Run("iphone_7.heic", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")

		assert.Nil(t, err)

		expectedBaseFilename := conf.ExamplesPath() + "/iphone_7"

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
	})
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

func TestMediaFileCanonicalName(t *testing.T) {
	conf := config.TestConfig()

	mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/beach_wood.jpg")
	assert.Nil(t, err)
	assert.Equal(t, "20180111_130938_EB4B2A989C20", mediaFile.CanonicalName())
}

func TestMediaFileCanonicalNameFromFile(t *testing.T) {
	conf := config.TestConfig()

	mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/beach_wood.jpg")
	assert.Nil(t, err)
	assert.Equal(t, "beach_wood", mediaFile.CanonicalNameFromFile())
}

func TestMediaFile_CanonicalNameFromFileWithDirectory(t *testing.T) {
	conf := config.TestConfig()

	mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/beach_wood.jpg")
	assert.Nil(t, err)
	assert.Equal(t, conf.ExamplesPath()+"/beach_wood", mediaFile.CanonicalNameFromFileWithDirectory())
}

func TestMediaFile_SetFilename(t *testing.T) {
	conf := config.TestConfig()

	mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/turtle_brown_blue.jpg")
	assert.Nil(t, err)
	mediaFile.SetFilename("newFilename")
	assert.Equal(t, "newFilename", mediaFile.filename)
	mediaFile.SetFilename("turtle_brown_blue")
	assert.Equal(t, "turtle_brown_blue", mediaFile.filename)
}

func TestMediaFile_Copy(t *testing.T) {
	conf := config.TestConfig()

	thumbsPath := conf.CachePath() + "/_tmp"

	os.MkdirAll(thumbsPath, os.ModePerm)

	defer os.RemoveAll(thumbsPath)

	mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/table_white.jpg")
	assert.Nil(t, err)
	mediaFile.Copy(thumbsPath + "table_whitecopy.jpg")
	assert.True(t, util.Exists(thumbsPath+"table_whitecopy.jpg"))

}

func TestMediaFile_DateCreated(t *testing.T) {
	conf := config.TestConfig()

	t.Run("iphone_7.heic", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic"); err == nil {

			info := mediaFile.CanonicalName()

			t.Log(info)

			exifinfo, err := mediaFile.Exif()

			t.Log(exifinfo.TakenAt)

			assert.Empty(t, err)
		} else {
			t.Error(err)
		}
	})
	t.Run("clock_purple.jpg", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/clock_purple.jpg"); err == nil {

			info := mediaFile.CanonicalName()

			t.Log(info)
			exifinfo, err := mediaFile.Exif()

			t.Log(exifinfo.TakenAt)

			assert.Empty(t, err)
		} else {
			t.Error(err)
		}
	})
	t.Run("canon_eos_6d.dng", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/canon_eos_6d.dng"); err == nil {

			info := mediaFile.DateCreated().UTC()

			exifinfo, err := mediaFile.Exif()

			t.Log(info)

			t.Log(exifinfo.TakenAt)

			assert.Empty(t, err)
		} else {
			t.Error(err)
		}
	})
	t.Run("elephants.jpg", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg"); err == nil {

			info := mediaFile.DateCreated().UTC()

			exifinfo, err := mediaFile.Exif()

			t.Log(info)

			t.Log(exifinfo.TakenAt)

			assert.Empty(t, err)
		} else {
			t.Error(err)
		}
	})
}
