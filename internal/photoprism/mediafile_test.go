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
	t.Run("/beach_wood.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/beach_wood.jpg")
		assert.Nil(t, err)
		assert.Equal(t, "beach_wood", mediaFile.CanonicalNameFromFile())
	})
	t.Run("/airport_grey", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/airport_grey")
		assert.Nil(t, err)
		assert.Equal(t, "airport_grey", mediaFile.CanonicalNameFromFile())
	})

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

func TestMediaFile_RelativeBasename(t *testing.T) {
	conf := config.TestConfig()

	mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/tree_white.jpg")
	assert.Nil(t, err)

	t.Run("directory with end slash", func(t *testing.T) {
		basename := mediaFile.RelativeBasename("/go/src/github.com/photoprism/photoprism/assets/resources/")
		assert.Equal(t, "examples/tree_white", basename)
	})
	t.Run("directory without end slash", func(t *testing.T) {
		basename := mediaFile.RelativeBasename("/go/src/github.com/photoprism/photoprism/assets/resources")
		assert.Equal(t, "examples/tree_white", basename)
	})
	t.Run("directory equals example path", func(t *testing.T) {
		basename := mediaFile.RelativeBasename("/go/src/github.com/photoprism/photoprism/assets/resources/examples/")
		assert.Equal(t, "tree_white", basename)
	})

}

func TestMediaFile_RelativePath(t *testing.T) {

	conf := config.TestConfig()

	mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/tree_white.jpg")
	assert.Nil(t, err)

	t.Run("directory with end slash", func(t *testing.T) {
		path := mediaFile.RelativePath("/go/src/github.com/photoprism/photoprism/assets/resources/")
		assert.Equal(t, "examples", path)
	})
	t.Run("directory without end slash", func(t *testing.T) {
		path := mediaFile.RelativePath("/go/src/github.com/photoprism/photoprism/assets/resources")
		assert.Equal(t, "examples", path)
	})
	t.Run("directory equals filepath", func(t *testing.T) {
		path := mediaFile.RelativePath("/go/src/github.com/photoprism/photoprism/assets/resources/examples")
		assert.Equal(t, "", path)
	})
	t.Run("directory does not match filepath", func(t *testing.T) {
		path := mediaFile.RelativePath("xxx")
		assert.Equal(t, "/go/src/github.com/photoprism/photoprism/assets/resources/examples", path)
	})
}

func TestMediaFile_RelativeFilename(t *testing.T) {
	conf := config.TestConfig()

	mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/tree_white.jpg")
	assert.Nil(t, err)

	t.Run("directory with end slash", func(t *testing.T) {
		filename := mediaFile.RelativeFilename("/go/src/github.com/photoprism/photoprism/assets/resources/")
		assert.Equal(t, "examples/tree_white.jpg", filename)
	})

	t.Run("directory without end slash", func(t *testing.T) {
		filename := mediaFile.RelativeFilename("/go/src/github.com/photoprism/photoprism/assets/resources")
		assert.Equal(t, "examples/tree_white.jpg", filename)
	})
	t.Run("directory not part of filename", func(t *testing.T) {
		filename := mediaFile.RelativeFilename("xxx/")
		assert.Equal(t, conf.ExamplesPath()+"/tree_white.jpg", filename)
	})
	t.Run("directory equals example path", func(t *testing.T) {
		filename := mediaFile.RelativeFilename("/go/src/github.com/photoprism/photoprism/assets/resources/examples")
		assert.Equal(t, "tree_white.jpg", filename)
	})
}

func TestMediaFile_DateCreated(t *testing.T) {
	conf := config.TestConfig()

	t.Run("iphone_7.heic", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic"); err == nil {
			date := mediaFile.DateCreated().UTC()
			assert.Equal(t, "2018-09-10 12:16:13 +0000 UTC", date.String())
			assert.Empty(t, err)
		} else {
			t.Error(err)
		}
	})
	t.Run("canon_eos_6d.dng", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/canon_eos_6d.dng"); err == nil {
			date := mediaFile.DateCreated().UTC()
			assert.Equal(t, "2019-06-06 07:29:51 +0000 UTC", date.String())
			assert.Empty(t, err)
		} else {
			t.Error(err)
		}
	})
	t.Run("elephants.jpg", func(t *testing.T) {
		if mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg"); err == nil {
			date := mediaFile.DateCreated().UTC()
			assert.Equal(t, "2013-11-26 15:53:55 +0000 UTC", date.String())
			assert.Empty(t, err)
		} else {
			t.Error(err)
		}
	})
}

func TestMediaFile_Height(t *testing.T) {
	t.Run("/Random.docx", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/Random.docx")
		assert.Nil(t, err)
		height := mediaFile.Height()
		assert.Equal(t, 0, height)
	})
	t.Run("/elephants.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")
		assert.Nil(t, err)
		height := mediaFile.Height()
		assert.Equal(t, 331, height)
	})
}

func TestMediaFile_AspectRatio(t *testing.T) {
	t.Run("/iphone_7.heic", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")
		assert.Nil(t, err)
		ratio := mediaFile.AspectRatio()
		assert.Equal(t, float64(0), ratio)
	})
	t.Run("/fern_green.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/fern_green.jpg")
		assert.Nil(t, err)
		ratio := mediaFile.AspectRatio()
		assert.Equal(t, float64(1), ratio)
	})
	t.Run("/elephants.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")
		assert.Nil(t, err)
		ratio := mediaFile.AspectRatio()
		assert.Equal(t, 1.501510574018127, ratio)
	})
}

func TestMediaFile_Orientation(t *testing.T) {
	t.Run("/iphone_7.heic", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")
		assert.Nil(t, err)
		orientation := mediaFile.Orientation()
		assert.Equal(t, 6, orientation)
	})
	t.Run("/turtle_brown_blue.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/turtle_brown_blue.jpg")
		assert.Nil(t, err)
		orientation := mediaFile.Orientation()
		assert.Equal(t, 1, orientation)
	})
}
