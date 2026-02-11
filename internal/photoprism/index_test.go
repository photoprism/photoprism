package photoprism

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestIndex_Start(t *testing.T) {
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

	indexOpt := IndexOptionsAll(cfg)
	indexOpt.Rescan = false

	found, updated := ind.Start(indexOpt)
	assert.GreaterOrEqual(t, len(found), 0)
	assert.GreaterOrEqual(t, updated, 0)

	t.Logf("index run 1: found %s", english.Plural(updated, "file", "files"))
	t.Logf("index run 1: updated %s", english.Plural(updated, "file", "files"))

	time.Sleep(time.Second)

	found, updated = ind.Start(indexOpt)
	assert.GreaterOrEqual(t, len(found), 0)
	assert.GreaterOrEqual(t, updated, 0)

	t.Logf("index run 2: found %s", english.Plural(updated, "file", "files"))
	t.Logf("index run 2: updated %s", english.Plural(updated, "file", "files"))

	time.Sleep(time.Second)

	found, updated = ind.Start(indexOpt)
	assert.GreaterOrEqual(t, len(found), 0)
	assert.GreaterOrEqual(t, updated, 0)

	t.Logf("index run 3: found %s", english.Plural(updated, "file", "files"))
	t.Logf("index run 3: updated %s", english.Plural(updated, "file", "files"))
}

func TestIndex_File(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cfg := config.TestConfig()
	initErr := cfg.InitializeTestData()
	assert.NoError(t, initErr)

	convert := NewConvert(cfg)
	ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())

	err := ind.FileName("xxx", IndexOptionsAll(cfg))

	assert.Equal(t, IndexFailed, err.Status)
}

func TestIndex_StartIgnoredMainFromRelatedFiles(t *testing.T) {
	cfg := config.NewMinimalTestConfigWithDb("index-ppignore-related", t.TempDir())

	oldCfg := Config()
	SetConfig(cfg)

	t.Cleanup(func() {
		_ = cfg.CloseDb()

		SetConfig(oldCfg)
		if oldCfg != nil {
			oldCfg.RegisterDb()
		}
	})

	ignoreName := filepath.Join(cfg.OriginalsPath(), fs.PPIgnoreFilename)
	if err := os.WriteFile(ignoreName, []byte("*.png\n"), fs.ModeFile); err != nil {
		t.Fatal(err)
	}

	pngData, err := os.ReadFile("testdata/photoprism.png")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(cfg.OriginalsPath(), "sample.png"), pngData, fs.ModeFile); err != nil {
		t.Fatal(err)
	}

	jpgData, err := os.ReadFile("testdata/2018-04-12 19_24_49.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(cfg.OriginalsPath(), "sample.png.jpg"), jpgData, fs.ModeFile); err != nil {
		t.Fatal(err)
	}

	ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())
	opt := NewIndexOptions("/", true, false, true, false, false, cfg)
	_, _ = ind.Start(opt)

	var file entity.File
	err = entity.UnscopedDb().
		First(&file, "file_root = ? AND file_name = ?", entity.RootOriginals, "sample.png").
		Error
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
