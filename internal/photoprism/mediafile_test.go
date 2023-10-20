package photoprism

import (
	"image"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestMediaFile_Ok(t *testing.T) {
	c := config.TestConfig()

	exists, err := NewMediaFile(c.ExamplesPath() + "/cat_black.jpg")

	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, exists.Ok())

	missing, err := NewMediaFile(c.ExamplesPath() + "/xxz.jpg")

	assert.NotNil(t, missing)
	assert.Error(t, err)
	assert.False(t, missing.Ok())
}

func TestMediaFile_Empty(t *testing.T) {
	c := config.TestConfig()

	exists, err := NewMediaFile(c.ExamplesPath() + "/cat_black.jpg")

	if err != nil {
		t.Fatal(err)
	}

	assert.False(t, exists.Empty())

	missing, err := NewMediaFile(c.ExamplesPath() + "/xxz.jpg")

	assert.NotNil(t, missing)
	assert.Error(t, err)
	assert.True(t, missing.Empty())
}

func TestMediaFile_DateCreated(t *testing.T) {
	conf := config.TestConfig()

	t.Run("telegram_2020-01-30_09-57-18.jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/telegram_2020-01-30_09-57-18.jpg")
		if err != nil {
			t.Fatal(err)
		}
		date := mediaFile.DateCreated().UTC()
		assert.Equal(t, "2020-01-30 09:57:18 +0000 UTC", date.String())
	})
	t.Run("Screenshot 2019-05-21 at 10.45.52.png", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/Screenshot 2019-05-21 at 10.45.52.png")
		if err != nil {
			t.Fatal(err)
		}
		date := mediaFile.DateCreated().UTC()
		assert.Equal(t, "2019-05-21 10:45:52 +0000 UTC", date.String())
	})
	t.Run("iphone_7.heic", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")
		if err != nil {
			t.Fatal(err)
		}
		date := mediaFile.DateCreated().UTC()
		assert.Equal(t, "2018-09-10 03:16:13 +0000 UTC", date.String())
	})
	t.Run("canon_eos_6d.dng", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/canon_eos_6d.dng")
		if err != nil {
			t.Fatal(err)
		}
		date := mediaFile.DateCreated().UTC()
		assert.Equal(t, "2019-06-06 07:29:51 +0000 UTC", date.String())
	})
	t.Run("elephants.jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")
		if err != nil {
			t.Fatal(err)
		}
		date := mediaFile.DateCreated().UTC()
		assert.Equal(t, "2013-11-26 13:53:55 +0000 UTC", date.String())
	})
	t.Run("dog_created_1919.jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/dog_created_1919.jpg")
		if err != nil {
			t.Fatal(err)
		}
		date := mediaFile.DateCreated().UTC()
		assert.Equal(t, "1919-05-04 05:59:26 +0000 UTC", date.String())
	})
}

func TestMediaFile_TakenAt(t *testing.T) {
	conf := config.TestConfig()
	t.Run("testdata/2018-04-12 19_24_49.gif", func(t *testing.T) {
		mediaFile, err := NewMediaFile("testdata/2018-04-12 19_24_49.gif")
		if err != nil {
			t.Fatal(err)
		}

		date, src := mediaFile.TakenAt()
		assert.Equal(t, "2018-04-12 19:24:49 +0000 UTC", date.String())
		assert.Equal(t, entity.SrcName, src)
	})
	t.Run("testdata/2018-04-12 19_24_49.jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile("testdata/2018-04-12 19_24_49.jpg")
		if err != nil {
			t.Fatal(err)
		}

		date, src := mediaFile.TakenAt()
		assert.Equal(t, "2018-04-12 19:24:49 +0000 UTC", date.String())
		assert.Equal(t, entity.SrcName, src)
	})
	t.Run("telegram_2020-01-30_09-57-18.jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/telegram_2020-01-30_09-57-18.jpg")
		if err != nil {
			t.Fatal(err)
		}

		date, src := mediaFile.TakenAt()
		assert.Equal(t, "2020-01-30 09:57:18 +0000 UTC", date.String())
		assert.Equal(t, entity.SrcName, src)
	})
	t.Run("Screenshot 2019-05-21 at 10.45.52.png", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/Screenshot 2019-05-21 at 10.45.52.png")
		if err != nil {
			t.Fatal(err)
		}

		date, src := mediaFile.TakenAt()
		assert.Equal(t, "2019-05-21 10:45:52 +0000 UTC", date.String())
		assert.Equal(t, entity.SrcName, src)
	})
	t.Run("iphone_7.heic", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")
		if err != nil {
			t.Fatal(err)
		}

		date, src := mediaFile.TakenAt()
		assert.Equal(t, "2018-09-10 03:16:13 +0000 UTC", date.String())
		assert.Equal(t, entity.SrcMeta, src)
	})
	t.Run("canon_eos_6d.dng", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/canon_eos_6d.dng")
		if err != nil {
			t.Fatal(err)
		}

		date, src := mediaFile.TakenAt()
		assert.Equal(t, "2019-06-06 07:29:51 +0000 UTC", date.String())
		assert.Equal(t, entity.SrcMeta, src)
	})
	t.Run("elephants.jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")
		if err != nil {
			t.Fatal(err)
		}

		date, src := mediaFile.TakenAt()
		assert.Equal(t, "2013-11-26 13:53:55 +0000 UTC", date.String())
		assert.Equal(t, entity.SrcMeta, src)
	})
	t.Run("dog_created_1919.jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/dog_created_1919.jpg")
		if err != nil {
			t.Fatal(err)
		}

		date, src := mediaFile.TakenAt()
		assert.Equal(t, "1919-05-04 05:59:26 +0000 UTC", date.String())
		assert.Equal(t, entity.SrcMeta, src)
	})
}

func TestMediaFile_HasTimeAndPlace(t *testing.T) {
	t.Run("/beach_wood.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/beach_wood.jpg")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, true, mediaFile.HasTimeAndPlace())
	})
	t.Run("/peacock_blue.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/peacock_blue.jpg")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, mediaFile.HasTimeAndPlace())
	})
}
func TestMediaFile_CameraModel(t *testing.T) {
	t.Run("/beach_wood.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/beach_wood.jpg")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "iPhone SE", mediaFile.CameraModel())
	})
	t.Run("/iphone_7.heic", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "iPhone 7", mediaFile.CameraModel())
	})
}

func TestMediaFile_CameraMake(t *testing.T) {
	t.Run("/beach_wood.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/beach_wood.jpg")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "Apple", mediaFile.CameraMake())
	})
	t.Run("/peacock_blue.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/peacock_blue.jpg")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "", mediaFile.CameraMake())
	})
}

func TestMediaFile_LensModel(t *testing.T) {
	t.Run("/beach_wood.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/beach_wood.jpg")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "iPhone SE back camera 4.15mm f/2.2", mediaFile.LensModel())
	})
	t.Run("/canon_eos_6d.dng", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/canon_eos_6d.dng")
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "EF24-105mm f/4L IS USM", mediaFile.LensModel())
	})
}

func TestMediaFile_LensMake(t *testing.T) {
	t.Run("/cat_brown.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/cat_brown.jpg")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "Apple", mediaFile.LensMake())
	})
	t.Run("/elephants.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "", mediaFile.LensMake())
	})
}

func TestMediaFile_FocalLength(t *testing.T) {
	c := config.TestConfig()

	t.Run("/cat_brown.jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile(c.ExamplesPath() + "/cat_brown.jpg")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 29, mediaFile.FocalLength())
	})
	t.Run("/elephants.jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile(c.ExamplesPath() + "/elephants.jpg")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 111, mediaFile.FocalLength())
	})
}

func TestMediaFile_FNumber(t *testing.T) {
	c := config.TestConfig()

	t.Run("/cat_brown.jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile(c.ExamplesPath() + "/cat_brown.jpg")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, float32(2.2), mediaFile.FNumber())
	})
	t.Run("/elephants.jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile(c.ExamplesPath() + "/elephants.jpg")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, float32(10.0), mediaFile.FNumber())
	})
}

func TestMediaFile_Iso(t *testing.T) {
	c := config.TestConfig()

	t.Run("/cat_brown.jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile(c.ExamplesPath() + "/cat_brown.jpg")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 32, mediaFile.Iso())
	})
	t.Run("/elephants.jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile(c.ExamplesPath() + "/elephants.jpg")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 200, mediaFile.Iso())
	})
}

func TestMediaFile_Exposure(t *testing.T) {
	c := config.TestConfig()

	t.Run("/cat_brown.jpg", func(t *testing.T) {

		mediaFile, err := NewMediaFile(c.ExamplesPath() + "/cat_brown.jpg")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "1/50", mediaFile.Exposure())
	})
	t.Run("/elephants.jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile(c.ExamplesPath() + "/elephants.jpg")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "1/640", mediaFile.Exposure())
	})
}

func TestMediaFileCanonicalName(t *testing.T) {
	c := config.TestConfig()

	mediaFile, err := NewMediaFile(c.ExamplesPath() + "/beach_wood.jpg")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "20180111_110938_7D8F8A23", mediaFile.CanonicalName())
}

func TestMediaFileCanonicalNameFromFile(t *testing.T) {
	t.Run("/beach_wood.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/beach_wood.jpg")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "beach_wood", mediaFile.CanonicalNameFromFile())
	})
	t.Run("/airport_grey", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/airport_grey")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "airport_grey", mediaFile.CanonicalNameFromFile())
	})
}

func TestMediaFile_CanonicalNameFromFileWithDirectory(t *testing.T) {
	c := config.TestConfig()

	mediaFile, err := NewMediaFile(c.ExamplesPath() + "/beach_wood.jpg")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, c.ExamplesPath()+"/beach_wood", mediaFile.CanonicalNameFromFileWithDirectory())
}

func TestMediaFile_EditedFilename(t *testing.T) {
	c := config.TestConfig()

	t.Run("IMG_4120.JPG", func(t *testing.T) {
		mediaFile, err := NewMediaFile(c.ExamplesPath() + "/IMG_4120.JPG")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, c.ExamplesPath()+"/IMG_E4120.JPG", mediaFile.EditedName())
	})
	t.Run("fern_green.jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile(c.ExamplesPath() + "/fern_green.jpg")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "", mediaFile.EditedName())
	})
}

func TestMediaFile_SetFilename(t *testing.T) {
	c := config.TestConfig()

	mediaFile, err := NewMediaFile(c.ExamplesPath() + "/turtle_brown_blue.jpg")
	if err != nil {
		t.Fatal(err)
	}
	mediaFile.SetFileName("newFilename")
	assert.Equal(t, "newFilename", mediaFile.fileName)
	mediaFile.SetFileName("turtle_brown_blue")
	assert.Equal(t, "turtle_brown_blue", mediaFile.fileName)
}

func TestMediaFile_RootRelName(t *testing.T) {
	conf := config.TestConfig()

	mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/tree_white.jpg")

	if err != nil {
		t.Fatal(err)
	}

	t.Run("examples_path", func(t *testing.T) {
		filename := mediaFile.RootRelName()
		assert.Equal(t, "tree_white.jpg", filename)
	})
}

func TestMediaFile_RootRelPath(t *testing.T) {
	conf := config.TestConfig()

	mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/tree_white.jpg")
	mediaFile.fileRoot = entity.RootImport
	if err != nil {
		t.Fatal(err)
	}

	t.Run("examples_path", func(t *testing.T) {
		path := mediaFile.RootRelPath()
		assert.Equal(t, conf.ExamplesPath(), path)
	})
}

func TestMediaFile_RootPath(t *testing.T) {
	conf := config.TestConfig()

	mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/tree_white.jpg")

	if err != nil {
		t.Fatal(err)
	}

	mediaFile.fileRoot = entity.RootImport
	t.Run("examples_path", func(t *testing.T) {
		path := mediaFile.RootPath()
		assert.Contains(t, path, "import")
	})
}

func TestMediaFile_RelName(t *testing.T) {
	conf := config.TestConfig()

	mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/tree_white.jpg")

	if err != nil {
		t.Fatal(err)
	}

	t.Run("directory with end slash", func(t *testing.T) {
		filename := mediaFile.RelName(conf.AssetsPath())
		assert.Equal(t, "examples/tree_white.jpg", filename)
	})

	t.Run("directory without end slash", func(t *testing.T) {
		filename := mediaFile.RelName(conf.AssetsPath())
		assert.Equal(t, "examples/tree_white.jpg", filename)
	})
	t.Run("directory not part of filename", func(t *testing.T) {
		filename := mediaFile.RelName("xxx/")
		assert.Equal(t, conf.ExamplesPath()+"/tree_white.jpg", filename)
	})
	t.Run("directory equals example path", func(t *testing.T) {
		filename := mediaFile.RelName(conf.ExamplesPath())
		assert.Equal(t, "tree_white.jpg", filename)
	})
}

func TestMediaFile_RelativePath(t *testing.T) {

	conf := config.TestConfig()

	mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/tree_white.jpg")

	if err != nil {
		t.Fatal(err)
	}

	t.Run("directory with end slash", func(t *testing.T) {
		path := mediaFile.RelPath(conf.AssetsPath() + "/")
		assert.Equal(t, "examples", path)
	})
	t.Run("directory without end slash", func(t *testing.T) {
		path := mediaFile.RelPath(conf.AssetsPath())
		assert.Equal(t, "examples", path)
	})
	t.Run("directory equals filepath", func(t *testing.T) {
		path := mediaFile.RelPath(conf.ExamplesPath())
		assert.Equal(t, "", path)
	})
	t.Run("directory does not match filepath", func(t *testing.T) {
		path := mediaFile.RelPath("xxx")
		assert.Equal(t, conf.ExamplesPath(), path)
	})

	mediaFile, err = NewMediaFile(conf.ExamplesPath() + "/.photoprism/example.jpg")

	if err != nil {
		t.Fatal(err)
	}

	t.Run("hidden", func(t *testing.T) {
		path := mediaFile.RelPath(conf.ExamplesPath())
		assert.Equal(t, "", path)
	})
	t.Run("hidden empty", func(t *testing.T) {
		path := mediaFile.RelPath("")
		assert.Equal(t, conf.ExamplesPath(), path)
	})
	t.Run("hidden root", func(t *testing.T) {
		path := mediaFile.RelPath(filepath.Join(conf.ExamplesPath(), fs.HiddenPath))
		assert.Equal(t, "", path)
	})
}

func TestMediaFile_RelativeBasename(t *testing.T) {
	conf := config.TestConfig()

	mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/tree_white.jpg")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("directory with end slash", func(t *testing.T) {
		basename := mediaFile.RelPrefix(conf.AssetsPath()+"/", true)
		assert.Equal(t, "examples/tree_white", basename)
	})
	t.Run("directory without end slash", func(t *testing.T) {
		basename := mediaFile.RelPrefix(conf.AssetsPath(), true)
		assert.Equal(t, "examples/tree_white", basename)
	})
	t.Run("directory equals example path", func(t *testing.T) {
		basename := mediaFile.RelPrefix(conf.ExamplesPath(), true)
		assert.Equal(t, "tree_white", basename)
	})

}

func TestMediaFile_Directory(t *testing.T) {
	t.Run("/limes.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/limes.jpg")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, conf.ExamplesPath(), mediaFile.Dir())
	})
}

func TestMediaFile_AbsPrefix(t *testing.T) {
	c := config.TestConfig()

	t.Run("/limes.jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile(c.ExamplesPath() + "/limes.jpg")
		if err != nil {
			t.Fatal(err)
		}

		expected := c.ExamplesPath() + "/limes"
		assert.Equal(t, expected, mediaFile.AbsPrefix(true))
	})
	t.Run("/IMG_4120 copy.JPG", func(t *testing.T) {
		mediaFile, err := NewMediaFile(c.ExamplesPath() + "/IMG_4120 copy.JPG")
		if err != nil {
			t.Fatal(err)
		}

		expected := c.ExamplesPath() + "/IMG_4120"
		assert.Equal(t, expected, mediaFile.AbsPrefix(true))
	})
	t.Run("/IMG_4120 (1).JPG", func(t *testing.T) {
		mediaFile, err := NewMediaFile(c.ExamplesPath() + "/IMG_4120 (1).JPG")
		if err != nil {
			t.Fatal(err)
		}

		expected := c.ExamplesPath() + "/IMG_4120"
		assert.Equal(t, expected, mediaFile.AbsPrefix(true))
	})
}

func TestMediaFile_BasePrefix(t *testing.T) {
	c := config.TestConfig()

	t.Run("/limes.jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile(c.ExamplesPath() + "/limes.jpg")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "limes", mediaFile.BasePrefix(true))
	})
	t.Run("/IMG_4120 copy.JPG", func(t *testing.T) {
		mediaFile, err := NewMediaFile(c.ExamplesPath() + "/IMG_4120 copy.JPG")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "IMG_4120", mediaFile.BasePrefix(true))
	})
	t.Run("/IMG_4120 (1).JPG", func(t *testing.T) {
		mediaFile, err := NewMediaFile(c.ExamplesPath() + "/IMG_4120 (1).JPG")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "IMG_4120", mediaFile.BasePrefix(true))
	})
}

func TestMediaFile_MimeType(t *testing.T) {
	conf := config.TestConfig()

	t.Run("elephants.jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "image/jpeg", mediaFile.MimeType())
	})

	t.Run("canon_eos_6d.dng", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/canon_eos_6d.dng")
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "image/dng", mediaFile.MimeType())
		assert.True(t, mediaFile.IsDNG())
		assert.True(t, mediaFile.IsRaw())
	})

	t.Run("iphone_7.xmp", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.xmp")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "text/plain", mediaFile.MimeType())
	})

	t.Run("iphone_7.json", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.json")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "application/json", mediaFile.MimeType())
	})
	t.Run("fox.profile0.8bpc.yuv420.avif", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/fox.profile0.8bpc.yuv420.avif")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, fs.MimeTypeAVIF, mediaFile.MimeType())
		assert.True(t, mediaFile.IsAVIF())
	})
	t.Run("iphone_7.heic", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, fs.MimeTypeHEIC, mediaFile.MimeType())
		assert.True(t, mediaFile.IsHEIC())
	})
	t.Run("IMG_4120.AAE", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/IMG_4120.AAE")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, fs.MimeTypeXML, mediaFile.MimeType())
	})

	t.Run("earth.mov", func(t *testing.T) {
		if f, err := NewMediaFile(filepath.Join(conf.ExamplesPath(), "earth.mov")); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "video/quicktime", f.MimeType())
		}
	})

	t.Run("blue-go-video.mp4", func(t *testing.T) {
		if f, err := NewMediaFile(filepath.Join(conf.ExamplesPath(), "blue-go-video.mp4")); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "video/mp4", f.MimeType())
		}
	})

	t.Run("earth.avi", func(t *testing.T) {
		if f, err := NewMediaFile(filepath.Join(conf.ExamplesPath(), "earth.avi")); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "video/x-msvideo", f.MimeType())
		}
	})

	t.Run("agpl.svg", func(t *testing.T) {
		if f, err := NewMediaFile("testdata/agpl.svg"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "image/svg+xml", f.MimeType())
		}
	})

	t.Run("favicon.ico", func(t *testing.T) {
		if f, err := NewMediaFile("testdata/favicon.ico"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "image/x-icon", f.MimeType())
		}
	})
}

func TestMediaFile_Exists(t *testing.T) {
	conf := config.TestConfig()

	exists, err := NewMediaFile(conf.ExamplesPath() + "/cat_black.jpg")

	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, exists)
	assert.True(t, exists.Exists())
	assert.Equal(t, true, exists.Ok())
	assert.Equal(t, false, exists.Empty())

	missing, err := NewMediaFile(conf.ExamplesPath() + "/xxz.jpg")

	assert.NotNil(t, missing)
	assert.Error(t, err)
	assert.Equal(t, int64(-1), missing.FileSize())
	assert.Equal(t, false, missing.Ok())
	assert.Equal(t, true, missing.Empty())
}

func TestMediaFile_Move(t *testing.T) {
	conf := config.TestConfig()

	tmpPath := conf.CachePath() + "/_tmp/TestMediaFile_Move"
	origName := tmpPath + "/original.jpg"
	destName := tmpPath + "/destination.jpg"

	if err := os.MkdirAll(tmpPath, fs.ModeDir); err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(tmpPath)

	f, err := NewMediaFile(conf.ExamplesPath() + "/table_white.jpg")

	if err != nil {
		t.Fatal(err)
	}

	if err := f.Copy(origName); err != nil {
		t.Fatal(err)
	}

	assert.True(t, fs.FileExists(origName))

	m, err := NewMediaFile(origName)
	if err != nil {
		t.Fatal(err)
	}

	if err = m.Move(destName); err != nil {
		t.Errorf("failed to move: %s", err)
	}

	assert.True(t, fs.FileExists(destName))
	assert.Equal(t, destName, m.FileName())
}

func TestMediaFile_Copy(t *testing.T) {
	conf := config.TestConfig()

	tmpPath := conf.CachePath() + "/_tmp/TestMediaFile_Copy"

	if err := os.MkdirAll(tmpPath, fs.ModeDir); err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(tmpPath)

	mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/table_white.jpg")

	if err != nil {
		t.Fatal(err)
	}

	if err := mediaFile.Copy(tmpPath + "table_whitecopy.jpg"); err != nil {
		t.Fatal(err)
	}

	assert.True(t, fs.FileExists(tmpPath+"table_whitecopy.jpg"))
}

func TestMediaFile_Extension(t *testing.T) {
	conf := config.TestConfig()

	t.Run("iphone_7.json", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.json")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, ".json", mediaFile.Extension())
	})
	t.Run("iphone_7.heic", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, ".heic", mediaFile.Extension())
	})
	t.Run("canon_eos_6d.dng", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/canon_eos_6d.dng")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, ".dng", mediaFile.Extension())
	})
	t.Run("elephants.jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, fs.ExtJPEG, mediaFile.Extension())
	})
}

func TestMediaFile_IsJpeg(t *testing.T) {
	conf := config.TestConfig()

	t.Run("iphone_7.json", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.json")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, mediaFile.IsJpeg())
	})
	t.Run("iphone_7.heic", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, mediaFile.IsJpeg())
	})
	t.Run("canon_eos_6d.dng", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/canon_eos_6d.dng")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, mediaFile.IsJpeg())
	})
	t.Run("elephants.jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, true, mediaFile.IsJpeg())
	})
}

func TestMediaFile_HasType(t *testing.T) {
	conf := config.TestConfig()

	t.Run("iphone_7.heic", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, mediaFile.HasFileType("jpg"))
	})
	t.Run("fox.profile0.8bpc.yuv420.avif", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/fox.profile0.8bpc.yuv420.avif")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, true, mediaFile.HasFileType("avif"))
	})
	t.Run("iphone_7.heic", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, true, mediaFile.HasFileType("heic"))
	})
	t.Run("iphone_7.xmp", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.xmp")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, true, mediaFile.HasFileType("xmp"))
	})
}

func TestMediaFile_IsHEIC(t *testing.T) {
	conf := config.TestConfig()

	t.Run("iphone_7.json", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.json")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, mediaFile.IsHEIC())
	})
	t.Run("iphone_7.heic", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, true, mediaFile.IsHEIC())
	})
	t.Run("canon_eos_6d.dng", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/canon_eos_6d.dng")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, mediaFile.IsHEIC())
	})
	t.Run("elephants.jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, mediaFile.IsHEIC())
	})
}

func TestMediaFile_IsRaw(t *testing.T) {
	conf := config.TestConfig()

	t.Run("iphone_7.json", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.json")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, mediaFile.IsRaw())
	})
	t.Run("iphone_7.heic", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, mediaFile.IsRaw())
	})
	t.Run("canon_eos_6d.dng", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/canon_eos_6d.dng")
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, true, mediaFile.IsRaw())
	})
	t.Run("elephants.jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, mediaFile.IsRaw())
	})
}

func TestMediaFile_IsPng(t *testing.T) {
	conf := config.TestConfig()

	t.Run("iphone_7.json", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.json")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, mediaFile.IsPNG())
	})
	t.Run("tweethog.png", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/tweethog.png")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, fs.ImagePNG, mediaFile.FileType())
		assert.Equal(t, "image/png", mediaFile.MimeType())
		assert.Equal(t, true, mediaFile.IsPNG())
	})
}

func TestMediaFile_IsTiff(t *testing.T) {
	conf := config.TestConfig()

	t.Run("iphone_7.json", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.json")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, fs.SidecarJSON, mediaFile.FileType())
		assert.Equal(t, fs.MimeTypeJSON, mediaFile.MimeType())
		assert.Equal(t, false, mediaFile.IsTIFF())
	})
	t.Run("purple.tiff", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/purple.tiff")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, fs.ImageTIFF, mediaFile.FileType())
		assert.Equal(t, "image/tiff", mediaFile.MimeType())
		assert.Equal(t, true, mediaFile.IsTIFF())
	})
	t.Run("example.tiff", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/example.tif")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, fs.ImageTIFF, mediaFile.FileType())
		assert.Equal(t, "image/tiff", mediaFile.MimeType())
		assert.Equal(t, true, mediaFile.IsTIFF())
	})
}

func TestMediaFile_IsImageOther(t *testing.T) {
	conf := config.TestConfig()

	t.Run("iphone_7.json", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.json")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, mediaFile.IsImageOther())
	})
	t.Run("purple.tiff", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/purple.tiff")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, true, mediaFile.IsImageOther())
	})
	t.Run("tweethog.png", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/tweethog.png")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, mediaFile.IsJpeg())
		assert.Equal(t, false, mediaFile.IsGIF())
		assert.Equal(t, true, mediaFile.IsPNG())
		assert.Equal(t, false, mediaFile.IsBMP())
		assert.Equal(t, false, mediaFile.IsWebP())
		assert.Equal(t, true, mediaFile.IsImage())
		assert.Equal(t, true, mediaFile.IsImageNative())
		assert.Equal(t, true, mediaFile.IsImageOther())
		assert.Equal(t, false, mediaFile.IsVideo())
		assert.Equal(t, true, mediaFile.SkipTranscoding())
	})
	t.Run("yellow_rose-small.bmp", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/yellow_rose-small.bmp")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, fs.ImageBMP, mediaFile.FileType())
		assert.Equal(t, "image/bmp", mediaFile.MimeType())
		assert.Equal(t, false, mediaFile.IsJpeg())
		assert.Equal(t, false, mediaFile.IsGIF())
		assert.Equal(t, true, mediaFile.IsBMP())
		assert.Equal(t, false, mediaFile.IsWebP())
		assert.Equal(t, true, mediaFile.IsImage())
		assert.Equal(t, true, mediaFile.IsImageNative())
		assert.Equal(t, true, mediaFile.IsImageOther())
		assert.Equal(t, false, mediaFile.IsVideo())
		assert.Equal(t, true, mediaFile.SkipTranscoding())
	})
	t.Run("preloader.gif", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/preloader.gif")
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, fs.ImageGIF, mediaFile.FileType())
		assert.Equal(t, "image/gif", mediaFile.MimeType())
		assert.Equal(t, false, mediaFile.IsJpeg())
		assert.Equal(t, true, mediaFile.IsGIF())
		assert.Equal(t, false, mediaFile.IsBMP())
		assert.Equal(t, false, mediaFile.IsWebP())
		assert.Equal(t, true, mediaFile.IsImage())
		assert.Equal(t, true, mediaFile.IsImageNative())
		assert.Equal(t, true, mediaFile.IsImageOther())
		assert.Equal(t, false, mediaFile.IsVideo())
		assert.Equal(t, true, mediaFile.SkipTranscoding())
	})
	t.Run("norway-kjetil-moe.webp", func(t *testing.T) {
		mediaFile, err := NewMediaFile("testdata/norway-kjetil-moe.webp")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, fs.ImageWebP, mediaFile.FileType())
		assert.Equal(t, fs.MimeTypeWebP, mediaFile.MimeType())
		assert.Equal(t, false, mediaFile.IsJpeg())
		assert.Equal(t, false, mediaFile.IsGIF())
		assert.Equal(t, false, mediaFile.IsBMP())
		assert.Equal(t, true, mediaFile.IsWebP())
		assert.Equal(t, true, mediaFile.IsImage())
		assert.Equal(t, true, mediaFile.IsImageNative())
		assert.Equal(t, true, mediaFile.IsImageOther())
		assert.Equal(t, false, mediaFile.IsVideo())
		assert.Equal(t, true, mediaFile.SkipTranscoding())
	})
}

func TestMediaFile_IsSidecar(t *testing.T) {
	conf := config.TestConfig()

	t.Run("iphone_7.xmp", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.xmp")
		assert.Nil(t, err)
		assert.Equal(t, true, mediaFile.IsSidecar())
	})
	t.Run("IMG_4120.AAE", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/IMG_4120.AAE")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, true, mediaFile.IsSidecar())
	})
	t.Run("test.xml", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/test.xml")

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, true, mediaFile.IsSidecar())
	})
	t.Run("test.txt", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/test.txt")

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, true, mediaFile.IsSidecar())
	})
	t.Run("test.yml", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/test.yml")

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, true, mediaFile.IsSidecar())
	})
	t.Run("test.md", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/test.md")

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, true, mediaFile.IsSidecar())
	})
	t.Run("canon_eos_6d.dng", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/canon_eos_6d.dng")

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, mediaFile.IsSidecar())
	})
}

func TestMediaFile_IsImage(t *testing.T) {
	cnf := config.TestConfig()

	t.Run("iphone_7.json", func(t *testing.T) {
		f, err := NewMediaFile(cnf.ExamplesPath() + "/iphone_7.json")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, f.IsImage())
		assert.Equal(t, false, f.IsRaw())
		assert.Equal(t, true, f.IsSidecar())
	})
	t.Run("iphone_7.xmp", func(t *testing.T) {
		f, err := NewMediaFile(cnf.ExamplesPath() + "/iphone_7.xmp")
		assert.Nil(t, err)
		assert.Equal(t, false, f.IsImage())
		assert.Equal(t, false, f.IsRaw())
		assert.Equal(t, true, f.IsSidecar())
	})
	t.Run("iphone_7.heic", func(t *testing.T) {
		f, err := NewMediaFile(cnf.ExamplesPath() + "/iphone_7.heic")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, true, f.IsImage())
		assert.Equal(t, false, f.IsRaw())
		assert.Equal(t, false, f.IsSidecar())
	})
	t.Run("canon_eos_6d.dng", func(t *testing.T) {
		f, err := NewMediaFile(cnf.ExamplesPath() + "/canon_eos_6d.dng")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, f.IsImage())
		assert.Equal(t, true, f.IsRaw())
		assert.Equal(t, false, f.IsSidecar())
	})
	t.Run("elephants.jpg", func(t *testing.T) {
		f, err := NewMediaFile(cnf.ExamplesPath() + "/elephants.jpg")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, true, f.IsImage())
		assert.Equal(t, false, f.IsRaw())
		assert.Equal(t, false, f.IsSidecar())
	})
}

func TestMediaFile_IsVideo(t *testing.T) {
	cnf := config.TestConfig()

	t.Run("christmas.mp4", func(t *testing.T) {
		if f, err := NewMediaFile(filepath.Join(cnf.ExamplesPath(), "christmas.mp4")); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, false, f.IsRaw())
			assert.Equal(t, false, f.IsImage())
			assert.Equal(t, true, f.IsVideo())
			assert.Equal(t, false, f.IsJSON())
			assert.Equal(t, false, f.IsSidecar())
		}
	})
	t.Run("canon_eos_6d.dng", func(t *testing.T) {
		if f, err := NewMediaFile(filepath.Join(cnf.ExamplesPath(), "canon_eos_6d.dng")); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, true, f.IsRaw())
			assert.Equal(t, false, f.IsImage())
			assert.Equal(t, false, f.IsVideo())
			assert.Equal(t, false, f.IsJSON())
			assert.Equal(t, false, f.IsSidecar())
		}
	})
	t.Run("iphone_7.json", func(t *testing.T) {
		if f, err := NewMediaFile(filepath.Join(cnf.ExamplesPath(), "iphone_7.json")); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, false, f.IsRaw())
			assert.Equal(t, false, f.IsImage())
			assert.Equal(t, false, f.IsVideo())
			assert.Equal(t, true, f.IsJSON())
			assert.Equal(t, true, f.IsSidecar())
		}
	})
}

func TestMediaFile_IsAnimated(t *testing.T) {
	cnf := config.TestConfig()
	t.Run("star.avifs", func(t *testing.T) {
		if f, err := NewMediaFile("testdata/star.avifs"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, true, f.IsImage())
			assert.Equal(t, true, f.IsAVIFS())
			assert.Equal(t, true, f.IsAnimated())
			assert.Equal(t, false, f.NotAnimated())
			assert.Equal(t, true, f.IsAnimatedImage())
			assert.Equal(t, true, f.ExifSupported())
			assert.Equal(t, false, f.IsVideo())
			assert.Equal(t, false, f.IsGIF())
			assert.Equal(t, false, f.IsWebP())
			assert.Equal(t, false, f.IsAVIF())
			assert.Equal(t, false, f.IsHEIC())
			assert.Equal(t, false, f.IsHEICS())
			assert.Equal(t, false, f.IsSidecar())
		}
	})
	t.Run("windows95.webp", func(t *testing.T) {
		if f, err := NewMediaFile("testdata/windows95.webp"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, true, f.IsImage())
			assert.Equal(t, true, f.IsWebP())
			assert.Equal(t, true, f.IsAnimated())
			assert.Equal(t, false, f.NotAnimated())
			assert.Equal(t, true, f.IsAnimatedImage())
			assert.Equal(t, false, f.ExifSupported())
			assert.Equal(t, false, f.IsVideo())
			assert.Equal(t, false, f.IsGIF())
			assert.Equal(t, false, f.IsAVIF())
			assert.Equal(t, false, f.IsAVIFS())
			assert.Equal(t, false, f.IsHEIC())
			assert.Equal(t, false, f.IsHEICS())
			assert.Equal(t, false, f.IsSidecar())
		}
	})
	t.Run("example.gif", func(t *testing.T) {
		if f, err := NewMediaFile(filepath.Join(cnf.ExamplesPath(), "example.gif")); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, true, f.IsImage())
			assert.Equal(t, false, f.IsVideo())
			assert.Equal(t, false, f.IsAnimated())
			assert.Equal(t, true, f.NotAnimated())
			assert.Equal(t, true, f.IsGIF())
			assert.Equal(t, false, f.IsAnimatedImage())
			assert.Equal(t, false, f.IsSidecar())
		}
	})
	t.Run("pythagoras.gif", func(t *testing.T) {
		if f, err := NewMediaFile(filepath.Join(cnf.ExamplesPath(), "pythagoras.gif")); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, true, f.IsImage())
			assert.Equal(t, false, f.IsVideo())
			assert.Equal(t, true, f.IsAnimated())
			assert.Equal(t, false, f.NotAnimated())
			assert.Equal(t, true, f.IsGIF())
			assert.Equal(t, true, f.IsAnimatedImage())
			assert.Equal(t, false, f.IsSidecar())
		}
	})
	t.Run("christmas.mp4", func(t *testing.T) {
		if f, err := NewMediaFile(filepath.Join(cnf.ExamplesPath(), "christmas.mp4")); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, false, f.IsImage())
			assert.Equal(t, true, f.IsVideo())
			assert.Equal(t, true, f.IsAnimated())
			assert.Equal(t, false, f.NotAnimated())
			assert.Equal(t, false, f.IsGIF())
			assert.Equal(t, false, f.IsAnimatedImage())
			assert.Equal(t, false, f.IsSidecar())
		}
	})
}

func TestMediaFile_HasPreviewImage(t *testing.T) {
	t.Run("Random.docx", func(t *testing.T) {
		cfg := config.TestConfig()

		f, err := NewMediaFile(cfg.ExamplesPath() + "/Random.docx")

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, f.HasPreviewImage())
	})
	t.Run("ferriswheel_colorful.jpg", func(t *testing.T) {
		cfg := config.TestConfig()

		f, err := NewMediaFile(cfg.ExamplesPath() + "/ferriswheel_colorful.jpg")

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, f.HasPreviewImage())
	})
	t.Run("Random.docx with jpg", func(t *testing.T) {
		cfg := config.TestConfig()

		f, err := NewMediaFile(cfg.ExamplesPath() + "/Random.docx")
		f.hasPreviewImage = true
		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, f.HasPreviewImage())
	})
}

func TestMediaFile_PreviewImage(t *testing.T) {
	t.Run("Random.docx", func(t *testing.T) {
		cfg := config.TestConfig()

		mediaFile, err := NewMediaFile(cfg.ExamplesPath() + "/Random.docx")

		if err != nil {
			t.Fatal(err)
		}

		file, err := mediaFile.PreviewImage()

		if file != nil {
			t.Fatal("file should be nil")
		}

		if err == nil {
			t.Fatal("err should NOT be nil")
		}

		assert.Equal(t, "no preview image found for Random.docx", err.Error())
	})
	t.Run("ferriswheel_colorful.jpg", func(t *testing.T) {
		cfg := config.TestConfig()

		mediaFile, err := NewMediaFile(cfg.ExamplesPath() + "/ferriswheel_colorful.jpg")

		if err != nil {
			t.Fatal(err)
		}

		file, err := mediaFile.PreviewImage()

		if err != nil {
			t.Fatal(err)
		}

		assert.FileExists(t, file.fileName)
	})
	t.Run("iphone_7.json", func(t *testing.T) {
		cfg := config.TestConfig()

		mediaFile, err := NewMediaFile(cfg.ExamplesPath() + "/test.md")

		if err != nil {
			t.Fatal(err)
		}

		file, err := mediaFile.PreviewImage()

		if file != nil {
			t.Fatal("file should be nil")
		}

		if err == nil {
			t.Fatal("err should NOT be nil")
		}

		assert.Equal(t, "no preview image found for test.md", err.Error())
	})
}

func TestMediaFile_decodeDimension(t *testing.T) {
	t.Run("Random.docx", func(t *testing.T) {
		cfg := config.TestConfig()

		mediaFile, err := NewMediaFile(cfg.ExamplesPath() + "/Random.docx")

		if err != nil {
			t.Fatal(err)
		}

		decodeErr := mediaFile.decodeDimensions()

		assert.EqualError(t, decodeErr, ".docx is not a valid media file")
	})

	t.Run("clock_purple.jpg", func(t *testing.T) {
		cfg := config.TestConfig()

		mediaFile, err := NewMediaFile(cfg.ExamplesPath() + "/clock_purple.jpg")

		if err != nil {
			t.Fatal(err)
		}

		if err := mediaFile.decodeDimensions(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("iphone_7.heic", func(t *testing.T) {
		cfg := config.TestConfig()

		mediaFile, err := NewMediaFile(cfg.ExamplesPath() + "/iphone_7.heic")

		if err != nil {
			t.Fatal(err)
		}

		if err := mediaFile.decodeDimensions(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("example.png", func(t *testing.T) {
		cfg := config.TestConfig()

		mediaFile, err := NewMediaFile(cfg.ExamplesPath() + "/example.png")

		if err != nil {
			t.Fatal(err)
		}

		if err := mediaFile.decodeDimensions(); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 100, mediaFile.Width())
		assert.Equal(t, 67, mediaFile.Height())
	})

	t.Run("example.gif", func(t *testing.T) {
		cfg := config.TestConfig()

		mediaFile, err := NewMediaFile(cfg.ExamplesPath() + "/example.gif")

		if err != nil {
			t.Fatal(err)
		}

		if err = mediaFile.decodeDimensions(); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 100, mediaFile.Width())
		assert.Equal(t, 67, mediaFile.Height())
	})

	t.Run("blue-go-video.mp4", func(t *testing.T) {
		cfg := config.TestConfig()

		mediaFile, err := NewMediaFile(cfg.ExamplesPath() + "/blue-go-video.mp4")

		if err != nil {
			t.Fatal(err)
		}

		if err = mediaFile.decodeDimensions(); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 1920, mediaFile.Width())
		assert.Equal(t, 1080, mediaFile.Height())
	})
	t.Run("blue-go-video.mp4 with orientation >4 and <8", func(t *testing.T) {
		cfg := config.TestConfig()

		mediaFile, err := NewMediaFile(cfg.ExamplesPath() + "/blue-go-video.mp4")
		mediaFile.metaData.Orientation = 5
		if err != nil {
			t.Fatal(err)
		}

		if err = mediaFile.decodeDimensions(); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 1080, mediaFile.Width())
		assert.Equal(t, 1920, mediaFile.Height())
	})
}

func TestMediaFile_Width(t *testing.T) {
	t.Run("Random.docx", func(t *testing.T) {
		cfg := config.TestConfig()

		mediaFile, err := NewMediaFile(cfg.ExamplesPath() + "/Random.docx")
		if err != nil {
			t.Fatal(err)
		}
		width := mediaFile.Width()
		assert.Equal(t, 0, width)
	})
	t.Run("elephant_mono.jpg", func(t *testing.T) {
		cfg := config.TestConfig()

		mediaFile, err := NewMediaFile(cfg.ExamplesPath() + "/elephant_mono.jpg")
		if err != nil {
			t.Fatal(err)
		}
		width := mediaFile.Width()
		assert.Equal(t, 416, width)
	})
}

func TestMediaFile_Height(t *testing.T) {
	t.Run("Random.docx", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/Random.docx")

		if err != nil {
			t.Fatal(err)
		}

		height := mediaFile.Height()
		assert.Equal(t, 0, height)
	})
	t.Run("elephants.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")

		if err != nil {
			t.Fatal(err)
		}

		height := mediaFile.Height()
		assert.Equal(t, 331, height)
	})
}

func TestMediaFile_Megapixels(t *testing.T) {
	conf := config.TestConfig()

	t.Run("Random.docx", func(t *testing.T) {
		if f, err := NewMediaFile(conf.ExamplesPath() + "/Random.docx"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, 0, f.Megapixels())
			assert.True(t, f.Ok())
			assert.False(t, f.Empty())
		}
	})
	t.Run("elephant_mono.jpg", func(t *testing.T) {
		if f, err := NewMediaFile(conf.ExamplesPath() + "/elephant_mono.jpg"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, 0, f.Megapixels())
			assert.True(t, f.Ok())
			assert.False(t, f.Empty())
		}
	})
	t.Run("telegram_2020-01-30_09-57-18.jpg", func(t *testing.T) {
		if f, err := NewMediaFile(conf.ExamplesPath() + "/telegram_2020-01-30_09-57-18.jpg"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, 1, f.Megapixels())
			assert.True(t, f.Ok())
			assert.False(t, f.Empty())
		}
	})
	t.Run("6720px_white.jpg", func(t *testing.T) {
		if f, err := NewMediaFile(conf.ExamplesPath() + "/6720px_white.jpg"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, 30, f.Megapixels())
			assert.True(t, f.Ok())
			assert.False(t, f.Empty())
		}
	})
	t.Run("canon_eos_6d.dng", func(t *testing.T) {
		if f, err := NewMediaFile(conf.ExamplesPath() + "/canon_eos_6d.dng"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, 0, f.Megapixels())
			assert.True(t, f.Ok())
			assert.False(t, f.Empty())
		}
	})
	t.Run("example.bmp", func(t *testing.T) {
		if f, err := NewMediaFile(conf.ExamplesPath() + "/example.bmp"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, 0, f.Megapixels())
			assert.True(t, f.Ok())
			assert.False(t, f.Empty())
		}
	})
	t.Run("panorama360.jpg", func(t *testing.T) {
		if f, err := NewMediaFile("testdata/panorama360.jpg"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, 0, f.Megapixels())
			assert.True(t, f.Ok())
			assert.False(t, f.Empty())
		}
	})
	t.Run("panorama360.json", func(t *testing.T) {
		if f, err := NewMediaFile("testdata/panorama360.json"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, 0, f.Megapixels())
			assert.True(t, f.Ok())
			assert.False(t, f.Empty())
		}
	})
	t.Run("2018-04-12 19_24_49.gif", func(t *testing.T) {
		if f, err := NewMediaFile("testdata/2018-04-12 19_24_49.gif"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, 0, f.Megapixels())
			assert.True(t, f.Ok())
			assert.False(t, f.Empty())
		}
	})
	t.Run("2018-04-12 19_24_49.mov", func(t *testing.T) {
		if f, err := NewMediaFile("testdata/2018-04-12 19_24_49.mov"); err != nil {
			t.Fatal(err)
		} else {
			assert.False(t, f.Ok())
			assert.True(t, f.Empty())
		}
	})
	t.Run("rotate/6.png", func(t *testing.T) {
		if f, err := NewMediaFile("testdata/rotate/6.png"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, 1, f.Megapixels())
			assert.True(t, f.Ok())
			assert.False(t, f.Empty())
		}
	})
	t.Run("rotate/6.tiff", func(t *testing.T) {
		if f, err := NewMediaFile("testdata/rotate/6.tiff"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, 0, f.Megapixels())
			assert.True(t, f.Ok())
			assert.False(t, f.Empty())
		}
	})
	t.Run("norway-kjetil-moe.webp", func(t *testing.T) {
		if f, err := NewMediaFile("testdata/norway-kjetil-moe.webp"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, 0, f.Megapixels())
			assert.True(t, f.Ok())
			assert.False(t, f.Empty())
		}
	})
}

func TestMediaFile_ExceedsBytes(t *testing.T) {
	t.Run("norway-kjetil-moe.webp", func(t *testing.T) {
		if f, err := NewMediaFile("testdata/norway-kjetil-moe.webp"); err != nil {
			t.Fatal(err)
		} else {
			err, actual := f.ExceedsBytes(3145728)
			assert.NoError(t, err)
			assert.Equal(t, int64(30320), actual)
			assert.True(t, f.Ok())
			assert.False(t, f.Empty())
		}
	})
	t.Run("telegram_2020-01-30_09-57-18.jpg", func(t *testing.T) {
		if f, err := NewMediaFile(conf.ExamplesPath() + "/telegram_2020-01-30_09-57-18.jpg"); err != nil {
			t.Fatal(err)
		} else {
			err, actual := f.ExceedsBytes(-1)
			assert.NoError(t, err)
			assert.Equal(t, int64(128471), actual)
			assert.True(t, f.Ok())
			assert.False(t, f.Empty())
		}
	})
	t.Run("6720px_white.jpg", func(t *testing.T) {
		if f, err := NewMediaFile(conf.ExamplesPath() + "/6720px_white.jpg"); err != nil {
			t.Fatal(err)
		} else {
			err, actual := f.ExceedsBytes(0)
			assert.NoError(t, err)
			assert.Equal(t, int64(162877), actual)
			assert.True(t, f.Ok())
			assert.False(t, f.Empty())
		}
	})
	t.Run("canon_eos_6d.dng", func(t *testing.T) {
		if f, err := NewMediaFile(conf.ExamplesPath() + "/canon_eos_6d.dng"); err != nil {
			t.Fatal(err)
		} else {
			err, actual := f.ExceedsBytes(10485760)
			assert.NoError(t, err)
			assert.Equal(t, int64(411944), actual)
			assert.True(t, f.Ok())
			assert.False(t, f.Empty())
		}
	})
	t.Run("example.bmp", func(t *testing.T) {
		if f, err := NewMediaFile(conf.ExamplesPath() + "/example.bmp"); err != nil {
			t.Fatal(err)
		} else {
			err, actual := f.ExceedsBytes(10485760)
			assert.NoError(t, err)
			assert.Equal(t, int64(20156), actual)
			assert.True(t, f.Ok())
			assert.False(t, f.Empty())
		}
	})
}
func TestMediaFile_DecodeConfig(t *testing.T) {
	t.Run("6720px_white.jpg", func(t *testing.T) {
		f, err := NewMediaFile(conf.ExamplesPath() + "/6720px_white.jpg")

		if err != nil {
			t.Fatal(err)
		}

		cfg1, err1 := f.DecodeConfig()

		assert.Nil(t, err1)
		assert.IsType(t, &image.Config{}, cfg1)
		assert.Equal(t, 6720, cfg1.Width)
		assert.Equal(t, 4480, cfg1.Height)

		cfg2, err2 := f.DecodeConfig()

		assert.Nil(t, err2)
		assert.IsType(t, &image.Config{}, cfg2)
		assert.Equal(t, 6720, cfg2.Width)
		assert.Equal(t, 4480, cfg2.Height)

		cfg3, err3 := f.DecodeConfig()

		assert.Nil(t, err3)
		assert.IsType(t, &image.Config{}, cfg3)
		assert.Equal(t, 6720, cfg3.Width)
		assert.Equal(t, 4480, cfg3.Height)
	})
}

func TestMediaFile_ExceedsResolution(t *testing.T) {
	t.Run("norway-kjetil-moe.webp", func(t *testing.T) {
		if f, err := NewMediaFile("testdata/norway-kjetil-moe.webp"); err != nil {
			t.Fatal(err)
		} else {
			result, actual := f.ExceedsResolution(3)
			assert.NoError(t, result)
			assert.Equal(t, 0, actual)
		}
	})
	t.Run("telegram_2020-01-30_09-57-18.jpg", func(t *testing.T) {
		if f, err := NewMediaFile(conf.ExamplesPath() + "/telegram_2020-01-30_09-57-18.jpg"); err != nil {
			t.Fatal(err)
		} else {
			result, actual := f.ExceedsResolution(3)
			assert.NoError(t, result)
			assert.Equal(t, 1, actual)
		}
	})
	t.Run("6720px_white.jpg", func(t *testing.T) {
		f, err := NewMediaFile(conf.ExamplesPath() + "/6720px_white.jpg")

		if err != nil {
			t.Fatal(err)
		}

		err3, actual3 := f.ExceedsResolution(3)

		assert.Error(t, err3)
		assert.Equal(t, 30, actual3)

		err30, actual30 := f.ExceedsResolution(30)

		assert.NoError(t, err30)
		assert.Equal(t, 30, actual30)

		err33, actual33 := f.ExceedsResolution(33)

		assert.NoError(t, err33)
		assert.Equal(t, 30, actual33)
	})
	t.Run("canon_eos_6d.dng", func(t *testing.T) {
		if f, err := NewMediaFile(conf.ExamplesPath() + "/canon_eos_6d.dng"); err != nil {
			t.Fatal(err)
		} else {
			result, actual := f.ExceedsResolution(3)
			assert.NoError(t, result)
			assert.Equal(t, 0, actual)
		}
	})
	t.Run("example.bmp", func(t *testing.T) {
		if f, err := NewMediaFile(conf.ExamplesPath() + "/example.bmp"); err != nil {
			t.Fatal(err)
		} else {
			result, actual := f.ExceedsResolution(3)
			assert.NoError(t, result)
			assert.Equal(t, 0, actual)
		}
	})
}

func TestMediaFile_AspectRatio(t *testing.T) {
	t.Run("iphone_7.heic", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")

		if err != nil {
			t.Fatal(err)
		}

		ratio := mediaFile.AspectRatio()
		assert.Equal(t, float32(0.75), ratio)
	})
	t.Run("fern_green.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/fern_green.jpg")

		if err != nil {
			t.Fatal(err)
		}

		ratio := mediaFile.AspectRatio()
		assert.Equal(t, float32(1), ratio)
	})
	t.Run("elephants.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")

		if err != nil {
			t.Fatal(err)
		}

		ratio := mediaFile.AspectRatio()
		assert.Equal(t, float32(1.5), ratio)
	})
}

func TestMediaFile_Orientation(t *testing.T) {
	t.Run("iphone_7.heic", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")

		if err != nil {
			t.Fatal(err)
		}

		orientation := mediaFile.Orientation()
		assert.Equal(t, 6, orientation)
	})
	t.Run("turtle_brown_blue.jpg", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/turtle_brown_blue.jpg")

		if err != nil {
			t.Fatal(err)
		}

		orientation := mediaFile.Orientation()
		assert.Equal(t, 1, orientation)
	})
}

func TestMediaFile_FileType(t *testing.T) {
	m, err := NewMediaFile(filepath.Join(conf.ExamplesPath(), "this-is-a-jpeg.png"))

	if err != nil {
		t.Fatal(err)
	}

	// No longer recognized as JPEG to improve indexing performance (skips mime type detection).
	assert.False(t, m.IsJpeg())
	assert.False(t, m.IsPNG())
	assert.Equal(t, "png", string(m.FileType()))
	assert.Equal(t, "image/jpeg", m.MimeType())
	assert.Equal(t, fs.ImagePNG, m.FileType())
	assert.Equal(t, ".png", m.Extension())
}

func TestMediaFile_Stat(t *testing.T) {
	t.Run("iphone_7.heic", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")

		if err != nil {
			t.Fatal(err)
		}

		size, time, err := mediaFile.Stat()

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, int64(785743), size)
		assert.IsType(t, time, time)
	})
}

func TestMediaFile_FileSize(t *testing.T) {
	t.Run("iphone_7.heic", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")

		if err != nil {
			t.Fatal(err)
		}

		size := mediaFile.FileSize()
		assert.Equal(t, int64(785743), size)
	})
}

func TestMediaFile_JsonName(t *testing.T) {
	t.Run("blue-go-video.mp4", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/blue-go-video.mp4")

		if err != nil {
			t.Fatal(err)
		}

		name := mediaFile.SidecarJsonName()
		assert.True(t, strings.HasSuffix(name, "/assets/examples/blue-go-video.mp4.json"))
	})
}

func TestMediaFile_PathNameInfo(t *testing.T) {
	t.Run("blue-go-video.mp4", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/blue-go-video.mp4")

		if err != nil {
			t.Fatal(err)
		}

		root, base, path, name := mediaFile.PathNameInfo(true)
		assert.Equal(t, "examples", root)
		assert.Equal(t, "blue-go-video", base)
		assert.Equal(t, "", path)
		assert.Equal(t, "blue-go-video.mp4", name)

	})

	t.Run("beach_sand sidecar", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/beach_sand.jpg")

		if err != nil {
			t.Fatal(err)
		}

		initialName := mediaFile.FileName()
		mediaFile.SetFileName(".photoprism/beach_sand.jpg")

		root, base, path, name := mediaFile.PathNameInfo(true)
		assert.Equal(t, "", root)
		assert.Equal(t, "beach_sand", base)
		assert.Equal(t, "", path)
		assert.Equal(t, ".photoprism/beach_sand.jpg", name)
		mediaFile.SetFileName(initialName)
	})

	t.Run("beach_sand import", func(t *testing.T) {
		conf := config.TestConfig()
		t.Log(Config().SidecarPath())
		t.Log(Config().ImportPath())

		mediaFile, err := NewMediaFile(filepath.Join(conf.ExamplesPath(), "beach_sand.jpg"))

		if err != nil {
			t.Fatal(err)
		}

		initialName := mediaFile.FileName()
		t.Log(initialName)
		mediaFile.SetFileName(filepath.Join(conf.ImportPath(), "beach_sand.jpg"))

		root, base, path, name := mediaFile.PathNameInfo(true)
		assert.Equal(t, "import", root)
		assert.Equal(t, "beach_sand", base)
		assert.Equal(t, "", path)
		assert.Equal(t, "beach_sand.jpg", name)
		mediaFile.SetFileName(initialName)
	})

	t.Run("beach_sand unknown root", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/beach_sand.jpg")

		if err != nil {
			t.Fatal(err)
		}

		initialName := mediaFile.FileName()
		mediaFile.SetFileName("/go/src/github.com/photoprism/notExisting/xxx/beach_sand.jpg")

		root, base, path, name := mediaFile.PathNameInfo(false)
		assert.Equal(t, "", root)
		assert.Equal(t, "beach_sand", base)
		assert.Equal(t, "/go/src/github.com/photoprism/notExisting/xxx", path)
		assert.Equal(t, "/go/src/github.com/photoprism/notExisting/xxx/beach_sand.jpg", name)
		mediaFile.SetFileName(initialName)
	})
}

func TestMediaFile_SubDirectory(t *testing.T) {
	t.Run("blue-go-video.mp4", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/blue-go-video.mp4")

		if err != nil {
			t.Fatal(err)
		}

		subdir := mediaFile.SubDir("xxx")
		assert.True(t, strings.HasSuffix(subdir, "/assets/examples/xxx"))
	})
}

func TestMediaFile_HasSameName(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/blue-go-video.mp4")

		if err != nil {
			t.Fatal(err)
		}

		mediaFile2, err := NewMediaFile(conf.ExamplesPath() + "/beach_sand.jpg")

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, mediaFile.HasSameName(nil))
		assert.False(t, mediaFile.HasSameName(mediaFile2))

	})
}

func TestMediaFile_IsJson(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/blue-go-video.mp4")

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, mediaFile.IsJSON())
	})
	t.Run("true", func(t *testing.T) {
		conf := config.TestConfig()

		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/blue-go-video.json")

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, mediaFile.IsJSON())
	})
}

func TestMediaFile_NeedsTranscoding(t *testing.T) {
	c := config.TestConfig()

	t.Run("json", func(t *testing.T) {
		f, err := NewMediaFile(c.ExamplesPath() + "/blue-go-video.json")

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, f.NeedsTranscoding())
	})
	t.Run("mp4", func(t *testing.T) {
		f, err := NewMediaFile(c.ExamplesPath() + "/blue-go-video.mp4")

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, f.NeedsTranscoding())
	})
	t.Run("mov", func(t *testing.T) {
		f, err := NewMediaFile(c.ExamplesPath() + "/earth.mov")

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, f.NeedsTranscoding())
	})
	t.Run("gif", func(t *testing.T) {
		f, err := NewMediaFile(c.ExamplesPath() + "/pythagoras.gif")

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, f.NeedsTranscoding())
	})
}

func TestMediaFile_SkipTranscoding(t *testing.T) {
	c := config.TestConfig()

	t.Run("json", func(t *testing.T) {
		f, err := NewMediaFile(c.ExamplesPath() + "/blue-go-video.json")

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, f.SkipTranscoding())
	})
	t.Run("mp4", func(t *testing.T) {
		f, err := NewMediaFile(c.ExamplesPath() + "/blue-go-video.mp4")

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, f.SkipTranscoding())
	})
	t.Run("mov", func(t *testing.T) {
		f, err := NewMediaFile(c.ExamplesPath() + "/earth.mov")

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, f.SkipTranscoding())
	})
	t.Run("gif", func(t *testing.T) {
		f, err := NewMediaFile(c.ExamplesPath() + "/pythagoras.gif")

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, f.SkipTranscoding())
	})
}

func TestMediaFile_RenameSidecarFiles(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		conf := config.TestConfig()

		jpegExample := filepath.Join(conf.ExamplesPath(), "/limes.jpg")
		jpegPath := filepath.Join(conf.OriginalsPath(), "2020", "12")
		jpegName := filepath.Join(jpegPath, "foobar.jpg")

		if err := fs.Copy(jpegExample, jpegName); err != nil {
			t.Fatal(err)
		}

		mf, err := NewMediaFile(jpegName)

		if err != nil {
			t.Fatal(err)
		}

		if err := os.MkdirAll(filepath.Join(conf.SidecarPath(), "foo"), fs.ModeDir); err != nil {
			t.Fatal(err)
		}

		srcName := filepath.Join(conf.SidecarPath(), "foo/bar.jpg.json")
		dstName := filepath.Join(conf.SidecarPath(), "2020/12/foobar.jpg.json")

		if err := os.WriteFile(srcName, []byte("{}"), 0666); err != nil {
			t.Fatal(err)
		}

		if renamed, err := mf.RenameSidecarFiles(filepath.Join(conf.OriginalsPath(), "foo/bar.jpg")); err != nil {
			t.Fatal(err)
		} else if len(renamed) != 1 {
			t.Errorf("len should be 2: %#v", renamed)
		} else {
			t.Logf("renamed: %#v", renamed)
		}

		if fs.FileExists(srcName) {
			t.Errorf("src file still exists: %s", srcName)
		}

		if !fs.FileExists(dstName) {
			t.Errorf("dst file not found: %s", srcName)
		}

		_ = os.Remove(srcName)
		_ = os.Remove(dstName)
	})
}

func TestMediaFile_RemoveSidecarFiles(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		conf := config.TestConfig()

		jpegExample := filepath.Join(conf.ExamplesPath(), "/limes.jpg")
		jpegPath := filepath.Join(conf.OriginalsPath(), "2020", "12")
		jpegName := filepath.Join(jpegPath, "foobar.jpg")

		if err := fs.Copy(jpegExample, jpegName); err != nil {
			t.Fatal(err)
		}

		mf, err := NewMediaFile(jpegName)

		if err != nil {
			t.Fatal(err)
		}

		sidecarName := filepath.Join(conf.SidecarPath(), "2020/12/foobar.jpg.json")

		if err := os.WriteFile(sidecarName, []byte("{}"), 0666); err != nil {
			t.Fatal(err)
		}

		if n, err := mf.RemoveSidecarFiles(); err != nil {
			t.Fatal(err)
		} else if fs.FileExists(sidecarName) {
			t.Errorf("src file still exists: %s", sidecarName)
		} else if n == 0 {
			t.Errorf("number of files should be > 0: %s", sidecarName)
		}

		_ = os.Remove(sidecarName)
	})
}

func TestMediaFile_ColorProfile(t *testing.T) {
	t.Run("iphone_7.json", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.json")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "", mediaFile.ColorProfile())
	})
	t.Run("iphone_7.xmp", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.xmp")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "", mediaFile.ColorProfile())
	})
	t.Run("iphone_7.heic", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/iphone_7.heic")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "", mediaFile.ColorProfile())
	})
	t.Run("canon_eos_6d.dng", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/canon_eos_6d.dng")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "", mediaFile.ColorProfile())
	})
	t.Run("elephants.jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/elephants.jpg")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "Adobe RGB (1998)", mediaFile.ColorProfile())
	})
	t.Run("/beach_wood.jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/beach_wood.jpg")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "", mediaFile.ColorProfile())
	})
	t.Run("/peacock_blue.jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile(conf.ExamplesPath() + "/peacock_blue.jpg")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "sRGB IEC61966-2.1", mediaFile.ColorProfile())
	})
}

func TestMediaFile_Duration(t *testing.T) {
	t.Run("earth.mov", func(t *testing.T) {
		if f, err := NewMediaFile(filepath.Join(conf.ExamplesPath(), "blue-go-video.mp4")); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "2.42s", f.Duration().String())
		}
	})
}
