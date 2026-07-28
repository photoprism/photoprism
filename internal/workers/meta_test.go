package workers

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestMeta_Start(t *testing.T) {
	conf := config.TestConfig()

	t.Logf("database-dsn: %s", conf.DatabaseDSN())

	worker := NewMeta(conf)

	assert.IsType(t, &Meta{}, worker)

	if err := mutex.MetaWorker.Start(); err != nil {
		t.Fatal(err)
	}

	delay := time.Second
	interval := time.Second

	// Mutex should prevent worker from starting.
	if err := worker.Start(delay, interval, true); err == nil {
		t.Fatal("error expected")
	}

	mutex.MetaWorker.Stop()

	// Start worker.
	if err := worker.Start(delay, interval, true); err != nil {
		t.Fatal(err)
	}

	// Rerun worker.
	if err := worker.Start(delay, interval, false); err != nil {
		t.Fatal(err)
	}
}

func TestMeta_originalsPath(t *testing.T) {
	conf := config.TestConfig()

	worker := NewMeta(conf)

	assert.IsType(t, &Meta{}, worker)
	assert.True(t, strings.HasSuffix(worker.originalsPath(), "testdata/originals"))
}

func TestMeta_saveSidecarYaml(t *testing.T) {
	// newPhoto returns an unsaved photo that marshals without a database round trip.
	newPhoto := func(name string) *entity.Photo {
		return &entity.Photo{
			PhotoUID:  rnd.GenerateUID(entity.PhotoUID),
			PhotoName: name,
			PhotoPath: "2026/07",
			PhotoType: entity.MediaImage,
		}
	}

	// yamlName returns the absolute sidecar backup path for a photo.
	yamlName := func(conf *config.Config, photo *entity.Photo) string {
		name, _, err := photo.YamlFileName(conf.OriginalsPath(), conf.SidecarPath())
		if err != nil {
			t.Fatal(err)
		}
		return name
	}

	t.Run("Success", func(t *testing.T) {
		conf := config.NewMinimalTestConfig("workers", t.TempDir())
		conf.Options().SidecarYaml = true
		conf.Options().DisableBackups = false
		photo := newPhoto("20260725-100000-success")
		NewMeta(conf).saveSidecarYaml(photo, "test")
		assert.FileExists(t, yamlName(conf, photo))
	})
	t.Run("Disabled", func(t *testing.T) {
		conf := config.NewMinimalTestConfig("workers", t.TempDir())
		conf.Options().SidecarYaml = false
		photo := newPhoto("20260725-100000-disabled")
		NewMeta(conf).saveSidecarYaml(photo, "test")
		assert.NoFileExists(t, yamlName(conf, photo))
	})
	t.Run("BackupsDisabled", func(t *testing.T) {
		conf := config.NewMinimalTestConfig("workers", t.TempDir())
		conf.Options().SidecarYaml = true
		conf.Options().DisableBackups = true
		photo := newPhoto("20260725-100000-nobackup")
		NewMeta(conf).saveSidecarYaml(photo, "test")
		assert.NoFileExists(t, yamlName(conf, photo))
	})
	t.Run("NilPhoto", func(t *testing.T) {
		conf := config.NewMinimalTestConfig("workers", t.TempDir())
		conf.Options().SidecarYaml = true
		conf.Options().DisableBackups = false
		assert.NotPanics(t, func() { NewMeta(conf).saveSidecarYaml(nil, "test") })
	})
	t.Run("InvalidPhoto", func(t *testing.T) {
		conf := config.NewMinimalTestConfig("workers", t.TempDir())
		conf.Options().SidecarYaml = true
		conf.Options().DisableBackups = false
		// A photo without a name cannot resolve a file name and must not panic or write.
		photo := &entity.Photo{PhotoUID: rnd.GenerateUID(entity.PhotoUID)}
		assert.NotPanics(t, func() { NewMeta(conf).saveSidecarYaml(photo, "test") })
	})
}
