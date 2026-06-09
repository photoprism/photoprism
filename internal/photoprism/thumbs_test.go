package photoprism

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestResample_Start(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cfg := config.TestConfig()

	if err := cfg.CreateDirectories(); err != nil {
		t.Fatal(err)
	}

	initErr := cfg.InitializeTestData()
	assert.NoError(t, initErr)

	convert := NewConvert(cfg)
	ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())

	imp := NewImport(cfg, ind, convert)
	opt := ImportOptionsMove(cfg.ImportPath(), "")

	imp.Start(opt)

	rs := NewThumbs(cfg)

	err := rs.Start("", true, false)

	if err != nil {
		t.Fatal(err)
	}
}

func TestThumbs_DirHonorsPPIgnore(t *testing.T) {
	cfg := config.NewMinimalTestConfig(t.TempDir())

	oldCfg := Config()
	SetConfig(cfg)

	t.Cleanup(func() {
		SetConfig(oldCfg)
	})

	dir := t.TempDir()

	ignoreName := filepath.Join(dir, fs.PPIgnoreFilename)
	if err := os.WriteFile(ignoreName, []byte("*.jpg\n"), fs.ModeFile); err != nil {
		t.Fatal(err)
	}

	jpgData, err := os.ReadFile("testdata/2018-04-12 19_24_49.jpg")
	if err != nil {
		t.Fatal(err)
	}

	imageName := filepath.Join(dir, "ignored.jpg")
	// The destination is a fixed test filename under t.TempDir().
	if err = os.WriteFile(imageName, jpgData, fs.ModeFile); err != nil { //nolint:gosec
		t.Fatal(err)
	}

	done, err := NewThumbs(cfg).Dir(dir, true)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := done[imageName]; !ok {
		t.Fatalf("expected %s to be tracked in done map", imageName)
	}

	assert.False(t, done[imageName].Processed())
}

func TestThumb_Filename(t *testing.T) {
	c := config.TestConfig()

	thumbsPath := c.CachePath() + "/_tmp"

	defer os.RemoveAll(thumbsPath)

	if err := c.CreateDirectories(); err != nil {
		t.Error(err)
	}

	t.Run("Success", func(t *testing.T) {
		filename, err := thumb.FileName("99988", thumbsPath, 150, 150, thumb.ResampleFit, thumb.ResampleNearestNeighbor)

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, strings.HasSuffix(filename, "/storage/testdata/cache/_tmp/9/9/9/99988_150x150_fit.jpg"))
	})
	t.Run("InvalidHash", func(t *testing.T) {
		_, err := thumb.FileName("999", thumbsPath, 150, 150, thumb.ResampleFit, thumb.ResampleNearestNeighbor)

		if err == nil {
			t.FailNow()
		}

		assert.Equal(t, "thumb: file hash is empty or too short (999)", err.Error())
	})
	t.Run("InvalidWidth", func(t *testing.T) {
		_, err := thumb.FileName("99988", thumbsPath, -4, 150, thumb.ResampleFit, thumb.ResampleNearestNeighbor)
		if err == nil {
			t.FailNow()
		}
		assert.Equal(t, "thumb: width exceeds limit (-4)", err.Error())
	})
	t.Run("InvalidHeight", func(t *testing.T) {
		_, err := thumb.FileName("99988", thumbsPath, 200, -1, thumb.ResampleFit, thumb.ResampleNearestNeighbor)
		if err == nil {
			t.FailNow()
		}
		assert.Equal(t, "thumb: height exceeds limit (-1)", err.Error())
	})
	t.Run("EmptyPath", func(t *testing.T) {
		path := ""
		_, err := thumb.FileName("99988", path, 200, 150, thumb.ResampleFit, thumb.ResampleNearestNeighbor)
		if err == nil {
			t.FailNow()
		}
		assert.Equal(t, "thumb: folder is empty", err.Error())
	})
}

func TestThumb_FromFile(t *testing.T) {
	c := config.TestConfig()

	thumbsPath := c.CachePath() + "/_tmp"

	defer os.RemoveAll(thumbsPath)

	if err := c.CreateDirectories(); err != nil {
		t.Error(err)
	}

	t.Run("ValidParameter", func(t *testing.T) {
		file := &entity.File{
			FileName: c.SamplesPath() + "/elephants.jpg",
			FileHash: "1234568889",
		}

		thumbnail, err := thumb.FromFile(file.FileName, file.FileHash, thumbsPath, 224, 224, file.FileOrientation)
		assert.Nil(t, err)
		assert.FileExists(t, thumbnail)
	})
	t.Run("HashTooShort", func(t *testing.T) {
		file := &entity.File{
			FileName: c.SamplesPath() + "/elephants.jpg",
			FileHash: "123",
		}

		_, err := thumb.FromFile(file.FileName, file.FileHash, thumbsPath, 224, 224, file.FileOrientation)

		if err == nil {
			t.Fatal("err should NOT be nil")
		}

		assert.Equal(t, "thumb: invalid file hash 123", err.Error())
	})
	t.Run("FilenameTooShort", func(t *testing.T) {
		file := &entity.File{
			FileName: "xxx",
			FileHash: "12367890",
		}

		if _, err := thumb.FromFile(file.FileName, file.FileHash, thumbsPath, 224, 224, file.FileOrientation); err != nil {
			assert.Equal(t, "thumb: invalid file name xxx", err.Error())
		} else {
			t.Error("error is nil")
		}
	})
	t.Run("RotateSixTiff", func(t *testing.T) {
		fileName := "testdata/rotate/6.tiff"

		file, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		thumbnail, err := thumb.FromFile(fileName, file.Hash(), thumbsPath, 224, 224, file.Orientation(), thumb.ResampleFit)

		if err != nil {
			t.Fatal(err)
		}

		tn, err := NewMediaFile(thumbnail)

		if err != nil {
			t.Fatal(err)
		}

		assert.NotNil(t, tn)
	})
}

func TestThumb_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	conf := config.TestConfig()

	thumbsPath := conf.CachePath() + "/_tmp"

	defer os.RemoveAll(thumbsPath)

	if err := conf.CreateDirectories(); err != nil {
		t.Error(err)
	}

	t.Run("ValidParameter", func(t *testing.T) {
		expectedFilename, err := thumb.FileName("12345", thumbsPath, 150, 150, thumb.ResampleFit, thumb.ResampleNearestNeighbor)

		if err != nil {
			t.Error(err)
		}

		img, _, err := fs.DecodeImageFile(conf.SamplesPath() + "/elephants.jpg")

		if err != nil {
			t.Errorf("cannot open original: %s", err)
		}

		res, err := thumb.Create(img, expectedFilename, 150, 150, thumb.ResampleFit, thumb.ResampleNearestNeighbor)

		if err != nil || res == nil {
			t.Fatal("err should be nil and res should NOT be nil")
		}

		thumbnail := res
		bounds := thumbnail.Bounds()

		assert.Equal(t, 150, bounds.Dx())
		assert.Equal(t, 99, bounds.Dy())

		assert.FileExists(t, expectedFilename)
	})
	t.Run("InvalidWidth", func(t *testing.T) {
		expectedFilename, err := thumb.FileName("12345", thumbsPath, 150, 150, thumb.ResampleFit, thumb.ResampleNearestNeighbor)

		if err != nil {
			t.Error(err)
		}

		img, _, err := fs.DecodeImageFile(conf.SamplesPath() + "/elephants.jpg")

		if err != nil {
			t.Errorf("cannot open original: %s", err)
		}

		res, err := thumb.Create(img, expectedFilename, -1, 150, thumb.ResampleFit, thumb.ResampleNearestNeighbor)

		if err == nil || res == nil {
			t.Fatal("err and res should NOT be nil")
		}

		thumbnail := res

		assert.Equal(t, "thumb: width has an invalid value (-1)", err.Error())
		bounds := thumbnail.Bounds()
		assert.NotEqual(t, 150, bounds.Dx())
	})
	t.Run("InvalidHeight", func(t *testing.T) {
		expectedFilename, err := thumb.FileName("12345", thumbsPath, 150, 150, thumb.ResampleFit, thumb.ResampleNearestNeighbor)

		if err != nil {
			t.Error(err)
		}

		img, _, err := fs.DecodeImageFile(conf.SamplesPath() + "/elephants.jpg")

		if err != nil {
			t.Errorf("cannot open original: %s", err)
		}

		res, err := thumb.Create(img, expectedFilename, 150, -1, thumb.ResampleFit, thumb.ResampleNearestNeighbor)

		if err == nil || res == nil {
			t.Fatal("err and res should NOT be nil")
		}

		thumbnail := res

		assert.Equal(t, "thumb: height has an invalid value (-1)", err.Error())
		bounds := thumbnail.Bounds()
		assert.NotEqual(t, 150, bounds.Dx())
	})
}
