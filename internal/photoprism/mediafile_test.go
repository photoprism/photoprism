package photoprism

import (
	"github.com/photoprism/photoprism/internal/util"
	"os"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestMediaFile_DateCreated(t *testing.T) {
	conf := config.TestConfig()

	t.Run("iphone_7.heic", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")
		assert.Nil(t, err)
		date := mediaFile.DateCreated().UTC()
		assert.Equal(t, "2018-09-10 03:16:13 +0000 UTC", date.String())
		assert.Empty(t, err)
	})
	t.Run("canon_eos_6d.dng", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/canon_eos_6d.dng")
		assert.Nil(t, err)
		date := mediaFile.DateCreated().UTC()
		assert.Equal(t, "2019-06-06 07:29:51 +0000 UTC", date.String())
		assert.Empty(t, err)
	})
	t.Run("elephants.jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")
		assert.Nil(t, err)
		date := mediaFile.DateCreated().UTC()
		assert.Equal(t, "2013-11-26 13:53:55 +0000 UTC", date.String())
		assert.Empty(t, err)
	})
}

func TestMediaFile_HasTimeAndPlace(t *testing.T) {
	t.Run("/beach_wood.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/beach_wood.jpg")
		assert.Nil(t, err)
		assert.Equal(t, true, mediaFile.HasTimeAndPlace())
	})
	t.Run("/peacock_blue.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/peacock_blue.jpg")
		assert.Nil(t, err)
		assert.Equal(t, false, mediaFile.HasTimeAndPlace())
	})
}
func TestMediaFile_CameraModel(t *testing.T) {
	t.Run("/beach_wood.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/beach_wood.jpg")
		assert.Nil(t, err)
		assert.Equal(t, "iPhone SE", mediaFile.CameraModel())
	})
	t.Run("/iphone_7.heic", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")
		assert.Nil(t, err)
		assert.Equal(t, "iPhone 7", mediaFile.CameraModel())
	})
}

func TestMediaFile_CameraMake(t *testing.T) {
	t.Run("/beach_wood.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/beach_wood.jpg")
		assert.Nil(t, err)
		assert.Equal(t, "Apple", mediaFile.CameraMake())
	})
	t.Run("/peacock_blue.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/peacock_blue.jpg")
		assert.Nil(t, err)
		assert.Equal(t, "", mediaFile.CameraMake())
	})
}

func TestMediaFile_LensModel(t *testing.T) {
	t.Run("/beach_wood.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/beach_wood.jpg")
		assert.Nil(t, err)
		assert.Equal(t, "iPhone SE back camera 4.15mm f/2.2", mediaFile.LensModel())
	})
	t.Run("/canon_eos_6d.dng", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/canon_eos_6d.dng")
		assert.Nil(t, err)
		assert.Equal(t, "EF24-105mm f/4L IS USM", mediaFile.LensModel())
	})
}

func TestMediaFile_LensMake(t *testing.T) {
	t.Run("/cat_brown.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/cat_brown.jpg")
		assert.Nil(t, err)
		assert.Equal(t, "Apple", mediaFile.LensMake())
	})
	t.Run("/elephants.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")
		assert.Nil(t, err)
		assert.Equal(t, "", mediaFile.LensMake())
	})
}

func TestMediaFile_FocalLength(t *testing.T) {
	t.Run("/cat_brown.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/cat_brown.jpg")
		assert.Nil(t, err)
		assert.Equal(t, 29, mediaFile.FocalLength())
	})
	t.Run("/elephants.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")
		assert.Nil(t, err)
		assert.Equal(t, 111, mediaFile.FocalLength())
	})
}

func TestMediaFile_FNumber(t *testing.T) {
	t.Run("/cat_brown.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/cat_brown.jpg")
		assert.Nil(t, err)
		assert.Equal(t, 2.2, mediaFile.FNumber())
	})
	t.Run("/elephants.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")
		assert.Nil(t, err)
		assert.Equal(t, 10.0, mediaFile.FNumber())
	})
}

func TestMediaFile_Iso(t *testing.T) {
	t.Run("/cat_brown.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/cat_brown.jpg")
		assert.Nil(t, err)
		assert.Equal(t, 32, mediaFile.Iso())
	})
	t.Run("/elephants.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")
		assert.Nil(t, err)
		assert.Equal(t, 200, mediaFile.Iso())
	})
}

func TestMediaFile_Exposure(t *testing.T) {
	t.Run("/cat_brown.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/cat_brown.jpg")
		assert.Nil(t, err)
		assert.Equal(t, "1/50", mediaFile.Exposure())
	})
	t.Run("/elephants.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")
		assert.Nil(t, err)
		assert.Equal(t, "1/640", mediaFile.Exposure())
	})
}

func TestMediaFileCanonicalName(t *testing.T) {
	conf := config.TestConfig()

	mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/beach_wood.jpg")
	assert.Nil(t, err)
	assert.Equal(t, "20180111_110938_EB4B2A989C20", mediaFile.CanonicalName())
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

func TestMediaFile_EditedFilename(t *testing.T) {
	conf := config.TestConfig()

	t.Run("IMG_4120.JPG", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/IMG_4120.JPG")
		assert.Nil(t, err)
		assert.Nil(t, err)
		assert.Equal(t, conf.ExamplesPath()+"/IMG_E4120.JPG", mediaFile.EditedFilename())
	})

	t.Run("fern_green.jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/fern_green.jpg")
		assert.Nil(t, err)
		assert.Nil(t, err)
		assert.Equal(t, "", mediaFile.EditedFilename())
	})
}

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

func TestMediaFile_SetFilename(t *testing.T) {
	conf := config.TestConfig()

	mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/turtle_brown_blue.jpg")
	assert.Nil(t, err)
	mediaFile.SetFilename("newFilename")
	assert.Equal(t, "newFilename", mediaFile.filename)
	mediaFile.SetFilename("turtle_brown_blue")
	assert.Equal(t, "turtle_brown_blue", mediaFile.filename)
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

func TestMediaFile_Directory(t *testing.T) {
	t.Run("/limes.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/limes.jpg")
		assert.Nil(t, err)
		assert.Equal(t, conf.ExamplesPath(), mediaFile.Directory())
	})
}

func TestMediaFile_Basename(t *testing.T) {
	t.Run("/limes.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/limes.jpg")
		assert.Nil(t, err)
		assert.Equal(t, "limes", mediaFile.Basename())
	})
	t.Run("/IMG_4120 copy.JPG", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/IMG_4120 copy.JPG")
		assert.Nil(t, err)
		assert.Equal(t, "IMG_4120", mediaFile.Basename())
	})
	t.Run("/IMG_4120 (1).JPG", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/IMG_4120 (1).JPG")
		assert.Nil(t, err)
		assert.Equal(t, "IMG_4120", mediaFile.Basename())
	})
}

func TestMediaFile_MimeType(t *testing.T) {
	conf := config.TestConfig()

	t.Run("elephants.jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")
		assert.Nil(t, err)
		assert.Nil(t, err)
		assert.Equal(t, "image/jpeg", mediaFile.MimeType())
	})

	t.Run("canon_eos_6d.dng", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/canon_eos_6d.dng")
		assert.Nil(t, err)
		assert.Nil(t, err)
		assert.Equal(t, "application/octet-stream", mediaFile.MimeType())

	})

	t.Run("iphone_7.xmp", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.xmp")
		assert.Nil(t, err)
		assert.Nil(t, err)
		assert.Equal(t, "text/plain; charset=utf-8", mediaFile.MimeType())
	})

	t.Run("iphone_7.json", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.json")
		assert.Nil(t, err)
		assert.Nil(t, err)
		assert.Equal(t, "text/plain; charset=utf-8", mediaFile.MimeType())
	})

	t.Run("iphone_7.heic", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")
		assert.Nil(t, err)
		assert.Nil(t, err)
		assert.Equal(t, "application/octet-stream", mediaFile.MimeType())
	})

	t.Run("IMG_4120.AAE", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/IMG_4120.AAE")
		assert.Nil(t, err)
		assert.Nil(t, err)
		assert.Equal(t, "text/xml; charset=utf-8", mediaFile.MimeType())
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

func TestMediaFile_Move(t *testing.T) {
	conf := config.TestConfig()

	tmpPath := conf.CachePath() + "/_tmp/TestMediaFile_Move"
	origName := tmpPath  + "/original.jpg"
	destName := tmpPath  + "/destination.jpg"

	os.MkdirAll(tmpPath, os.ModePerm)

	defer os.RemoveAll(tmpPath)

	f, err := NewMediaFile(conf.ExamplesPath() + "/table_white.jpg")
	assert.Nil(t, err)
	f.Copy(origName)
	assert.True(t, util.Exists(origName))

	m, err := NewMediaFile(origName)
	assert.Nil(t, err)

	if err = m.Move(destName); err != nil {
		t.Errorf("failed to move: %s", err)
	}

	assert.True(t, util.Exists(destName))
	assert.Equal(t, destName, m.Filename())
}

func TestMediaFile_Copy(t *testing.T) {
	conf := config.TestConfig()

	tmpPath := conf.CachePath() + "/_tmp/TestMediaFile_Copy"

	os.MkdirAll(tmpPath, os.ModePerm)

	defer os.RemoveAll(tmpPath)

	mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/table_white.jpg")
	assert.Nil(t, err)
	mediaFile.Copy(tmpPath + "table_whitecopy.jpg")
	assert.True(t, util.Exists(tmpPath+"table_whitecopy.jpg"))
}

func TestMediaFile_Extension(t *testing.T) {
	t.Run("/iphone_7.json", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.json")
		assert.Nil(t, err)
		assert.Equal(t, ".json", mediaFile.Extension())
	})
	t.Run("/iphone_7.heic", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")
		assert.Nil(t, err)
		assert.Equal(t, ".heic", mediaFile.Extension())
	})
	t.Run("/canon_eos_6d.dng", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/canon_eos_6d.dng")
		assert.Nil(t, err)
		assert.Equal(t, ".dng", mediaFile.Extension())
	})
	t.Run("/elephants.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")
		assert.Nil(t, err)
		assert.Equal(t, ".jpg", mediaFile.Extension())
	})
}

func TestMediaFile_IsJpeg(t *testing.T) {
	t.Run("/iphone_7.json", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.json")
		assert.Nil(t, err)
		assert.Equal(t, false, mediaFile.IsJpeg())
	})
	t.Run("/iphone_7.heic", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")
		assert.Nil(t, err)
		assert.Equal(t, false, mediaFile.IsJpeg())
	})
	t.Run("/canon_eos_6d.dng", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/canon_eos_6d.dng")
		assert.Nil(t, err)
		assert.Equal(t, false, mediaFile.IsJpeg())
	})
	t.Run("/elephants.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")
		assert.Nil(t, err)
		assert.Equal(t, true, mediaFile.IsJpeg())
	})
}

func TestMediaFile_HasType(t *testing.T) {
	t.Run("/iphone_7.heic", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")
		assert.Nil(t, err)
		assert.Equal(t, false, mediaFile.HasType("jpg"))
	})
	t.Run("/iphone_7.heic", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")
		assert.Nil(t, err)
		assert.Equal(t, true, mediaFile.HasType("heif"))
	})
	t.Run("/iphone_7.xmp", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.xmp")
		assert.Nil(t, err)
		assert.Equal(t, true, mediaFile.HasType("xmp"))
	})
}

func TestMediaFile_IsHEIF(t *testing.T) {
	t.Run("/iphone_7.json", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.json")
		assert.Nil(t, err)
		assert.Equal(t, false, mediaFile.IsHEIF())
	})
	t.Run("/iphone_7.heic", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")
		assert.Nil(t, err)
		assert.Equal(t, true, mediaFile.IsHEIF())
	})
	t.Run("/canon_eos_6d.dng", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/canon_eos_6d.dng")
		assert.Nil(t, err)
		assert.Equal(t, false, mediaFile.IsHEIF())
	})
	t.Run("/elephants.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")
		assert.Nil(t, err)
		assert.Equal(t, false, mediaFile.IsHEIF())
	})
}

func TestMediaFile_IsRaw(t *testing.T) {
	t.Run("/iphone_7.json", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.json")
		assert.Nil(t, err)
		assert.Equal(t, false, mediaFile.IsRaw())
	})
	t.Run("/iphone_7.heic", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")
		assert.Nil(t, err)
		assert.Equal(t, false, mediaFile.IsRaw())
	})
	t.Run("/canon_eos_6d.dng", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/canon_eos_6d.dng")
		assert.Nil(t, err)
		assert.Equal(t, true, mediaFile.IsRaw())
	})
	t.Run("/elephants.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")
		assert.Nil(t, err)
		assert.Equal(t, false, mediaFile.IsRaw())
	})
}

func TestMediaFile_IsPhoto(t *testing.T) {
	t.Run("/iphone_7.json", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.json")
		assert.Nil(t, err)
		assert.Equal(t, false, mediaFile.IsPhoto())
	})
	t.Run("/iphone_7.xmp", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.xmp")
		assert.Nil(t, err)
		assert.Equal(t, false, mediaFile.IsPhoto())
	})
	t.Run("/iphone_7.heic", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")
		assert.Nil(t, err)
		assert.Equal(t, true, mediaFile.IsPhoto())
	})
	t.Run("/canon_eos_6d.dng", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/canon_eos_6d.dng")
		assert.Nil(t, err)
		assert.Equal(t, true, mediaFile.IsPhoto())
	})
	t.Run("/elephants.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")
		assert.Nil(t, err)
		assert.Equal(t, true, mediaFile.IsPhoto())
	})
}

func TestMediaFile_Jpeg(t *testing.T) {
	t.Run("/Random.docx", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/Random.docx")
		assert.Nil(t, err)
		file, err := mediaFile.Jpeg()
		assert.Nil(t, file)
		assert.Equal(t, "jpeg file does not exist: "+conf.ExamplesPath()+"/Random.jpg", err.Error())
	})
	t.Run("/ferriswheel_colorful.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/ferriswheel_colorful.jpg")
		assert.Nil(t, err)
		file, err := mediaFile.Jpeg()
		assert.Nil(t, err)
		assert.FileExists(t, file.filename)
	})
	t.Run("/iphone_7.json", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.json")
		assert.Nil(t, err)
		file, err := mediaFile.Jpeg()
		assert.Nil(t, file)
		assert.Equal(t, "jpeg file does not exist: "+conf.ExamplesPath()+"/iphone_7.jpg", err.Error())
	})
}

func TestMediaFile_decodeDimension(t *testing.T) {
	t.Run("/Random.docx", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/Random.docx")
		assert.Nil(t, err)
		decodeErr := mediaFile.decodeDimensions()
		assert.Equal(t, "not a photo: "+conf.ExamplesPath()+"/Random.docx", decodeErr.Error())
	})
	t.Run("/clock_purple.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/clock_purple.jpg")
		assert.Nil(t, err)
		decodeErr := mediaFile.decodeDimensions()
		assert.Nil(t, decodeErr)
	})
	t.Run("/iphone_7.heic", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")
		assert.Nil(t, err)
		decodeErr := mediaFile.decodeDimensions()
		assert.Nil(t, decodeErr)
	})
}

func TestMediaFile_Width(t *testing.T) {
	t.Run("/Random.docx", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/Random.docx")
		assert.Nil(t, err)
		width := mediaFile.Width()
		assert.Equal(t, 0, width)
	})
	t.Run("/elephant_mono.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/elephant_mono.jpg")
		assert.Nil(t, err)
		width := mediaFile.Width()
		assert.Equal(t, 416, width)
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
