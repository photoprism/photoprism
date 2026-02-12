package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestNewImport(t *testing.T) {
	cfg := config.TestConfig()

	convert := NewConvert(cfg)

	ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())
	imp := NewImport(cfg, ind, convert)

	assert.IsType(t, &Import{}, imp)
}

func TestImport_DestinationFilename(t *testing.T) {
	cfg := config.TestConfig()

	initErr := cfg.InitializeTestData()
	assert.NoError(t, initErr)

	convert := NewConvert(cfg)

	ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())

	imp := NewImport(cfg, ind, convert)

	rawFile, err := NewMediaFile(cfg.ImportPath() + "/raw/IMG_2567.CR2")

	if err != nil {
		t.Fatal(err)
	}

	t.Run("NoBasePath", func(t *testing.T) {
		fileName, err := imp.DestinationFilename(rawFile, rawFile, "")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, cfg.OriginalsPath()+"/2019/07/20190705_153230_C167C6FD.cr2", fileName)
	})
	t.Run("WithBasePath", func(t *testing.T) {
		fileName, err := imp.DestinationFilename(rawFile, rawFile, "users/guest")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, cfg.OriginalsPath()+"/users/guest/2019/07/20190705_153230_C167C6FD.cr2", fileName)
	})
}

func TestImport_Start(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cfg := config.TestConfig()

	initErr := cfg.InitializeTestData()
	assert.NoError(t, initErr)

	convert := NewConvert(cfg)

	ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())

	imp := NewImport(cfg, ind, convert)

	opt := ImportOptionsMove(cfg.ImportPath(), "")

	imp.Start(opt)
}

func TestImport_StartIgnoredMainFromRelatedFiles(t *testing.T) {
	cfg := config.NewMinimalTestConfigWithDb("import-ppignore-related", t.TempDir())

	oldCfg := Config()
	SetConfig(cfg)

	t.Cleanup(func() {
		_ = cfg.CloseDb()

		SetConfig(oldCfg)
		if oldCfg != nil {
			oldCfg.RegisterDb()
		}
	})

	ignoreName := filepath.Join(cfg.ImportPath(), fs.PPIgnoreFilename)
	if err := os.WriteFile(ignoreName, []byte("*.png\n"), fs.ModeFile); err != nil {
		t.Fatal(err)
	}

	pngData, err := os.ReadFile("testdata/photoprism.png")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(cfg.ImportPath(), "sample.png"), pngData, fs.ModeFile); err != nil {
		t.Fatal(err)
	}

	jpgData, err := os.ReadFile("testdata/2018-04-12 19_24_49.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(cfg.ImportPath(), "sample.png.jpg"), jpgData, fs.ModeFile); err != nil {
		t.Fatal(err)
	}

	convert := NewConvert(cfg)
	ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())
	imp := NewImport(cfg, ind, convert)

	_ = imp.Start(ImportOptionsMove(cfg.ImportPath(), ""))

	var file entity.File
	err = entity.UnscopedDb().
		First(&file, "file_root = ? AND file_name = ?", entity.RootOriginals, "sample.png").
		Error
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	assert.True(t, fs.FileExists(filepath.Join(cfg.ImportPath(), "sample.png")))
	assert.True(t, fs.FileExists(filepath.Join(cfg.ImportPath(), "sample.png.jpg")))
}
