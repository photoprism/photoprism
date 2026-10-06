package workers

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
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

func TestMeta_Start_PublishesSavedPhotos(t *testing.T) {
	conf := config.NewMinimalTestConfigWithDb("workers-meta-events", t.TempDir())
	fixture := entity.PhotoFixtures.Get("VisionResetTarget")

	// Clearing the generated title makes the worker regenerate and save it.
	require.NoError(t, entity.Db().Model(&entity.Photo{}).Where("photo_uid = ?", fixture.PhotoUID).UpdateColumns(entity.Values{
		"PhotoTitle": "",
		"TitleSrc":   entity.SrcAuto,
		"CheckedAt":  nil,
		"UpdatedAt":  time.Now().Add(-time.Hour),
	}).Error)

	sub := subscribePhotoEvents(t, event.EntityUpdated)
	summary := captureEventsAtLog(t, sub, func(msg string) bool {
		return strings.HasPrefix(msg, "index: updated ") && !strings.HasPrefix(msg, "index: updated photo ")
	})

	require.NoError(t, NewMeta(conf).Start(time.Second, time.Second, false))

	// Every event is published when the loop ends, before face recognition and the index updates run.
	events := receivePhotoEvents(t, sub)
	require.True(t, summary.seen)
	assert.Equal(t, len(events), summary.count)

	var uids []string

	for _, ev := range events {
		assert.LessOrEqual(t, len(ev), event.EntityBatchSize)
		uids = append(uids, ev...)
	}

	assert.Contains(t, uids, fixture.PhotoUID)

	refreshed := entity.FindPhoto(fixture)
	require.NotNil(t, refreshed)
	assert.NotEmpty(t, refreshed.PhotoTitle)
}

func TestMeta_Start_PublishesMergedPhotos(t *testing.T) {
	conf := config.NewMinimalTestConfigWithDb("workers-meta-merge", t.TempDir())

	// newPhoto creates a stackable photo that the worker merges with others of the same name.
	newPhoto := func(quality int) *entity.Photo {
		photo := &entity.Photo{
			PhotoUID:     rnd.GenerateUID(entity.PhotoUID),
			PhotoType:    entity.MediaImage,
			PhotoPath:    "2026/10",
			PhotoName:    "merge-events",
			PhotoTitle:   "Merge Events",
			TitleSrc:     entity.SrcManual,
			PhotoQuality: quality,
			TakenAt:      time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC),
			TakenAtLocal: time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC),
			TakenSrc:     entity.SrcMeta,
			PlaceID:      entity.UnknownPlace.ID,
			CellID:       entity.UnknownLocation.ID,
		}
		require.NoError(t, entity.Db().Create(photo).Error)
		require.NoError(t, entity.Db().Model(photo).UpdateColumns(entity.Values{"CheckedAt": nil, "UpdatedAt": time.Now().Add(-time.Hour)}).Error)
		return photo
	}

	// The original keeps the higher quality score, so the duplicate is merged into it.
	original := newPhoto(4)

	// Optimizing the original up front leaves nothing to save, so only the merge can report it.
	settings := conf.Settings()
	for range 2 {
		_, _, err := original.Optimize(settings.StackMeta(), settings.StackUUID(), settings.Features.Estimates, false)
		require.NoError(t, err)
	}
	require.NoError(t, entity.Db().Model(original).UpdateColumns(entity.Values{"CheckedAt": nil, "UpdatedAt": time.Now().Add(-time.Hour)}).Error)

	duplicate := newPhoto(1)

	updated := subscribePhotoEvents(t, event.EntityUpdated)
	deleted := subscribePhotoEvents(t, event.EntityDeleted)

	require.NoError(t, NewMeta(conf).Start(time.Second, time.Second, false))

	var updatedUids, deletedUids []string

	for _, ev := range receivePhotoEvents(t, updated) {
		updatedUids = append(updatedUids, ev...)
	}

	for _, ev := range receivePhotoEvents(t, deleted) {
		deletedUids = append(deletedUids, ev...)
	}

	// Other fixtures may be merged in the same pass, so only the photos created here are checked.
	assert.Contains(t, updatedUids, original.PhotoUID)
	assert.NotContains(t, updatedUids, duplicate.PhotoUID)
	assert.Contains(t, deletedUids, duplicate.PhotoUID)
	assert.NotContains(t, deletedUids, original.PhotoUID)

	merged := entity.FindPhoto(*duplicate)
	require.NotNil(t, merged)
	assert.NotNil(t, merged.DeletedAt)
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
		conf := config.NewMinimalTestConfig(t.TempDir())
		conf.Options().SidecarYaml = true
		conf.Options().DisableBackups = false
		photo := newPhoto("20260725-100000-success")
		NewMeta(conf).saveSidecarYaml(photo, "test")
		assert.FileExists(t, yamlName(conf, photo))
	})
	t.Run("Disabled", func(t *testing.T) {
		conf := config.NewMinimalTestConfig(t.TempDir())
		conf.Options().SidecarYaml = false
		photo := newPhoto("20260725-100000-disabled")
		NewMeta(conf).saveSidecarYaml(photo, "test")
		assert.NoFileExists(t, yamlName(conf, photo))
	})
	t.Run("BackupsDisabled", func(t *testing.T) {
		conf := config.NewMinimalTestConfig(t.TempDir())
		conf.Options().SidecarYaml = true
		conf.Options().DisableBackups = true
		photo := newPhoto("20260725-100000-nobackup")
		NewMeta(conf).saveSidecarYaml(photo, "test")
		assert.NoFileExists(t, yamlName(conf, photo))
	})
	t.Run("NilPhoto", func(t *testing.T) {
		conf := config.NewMinimalTestConfig(t.TempDir())
		conf.Options().SidecarYaml = true
		conf.Options().DisableBackups = false
		assert.NotPanics(t, func() { NewMeta(conf).saveSidecarYaml(nil, "test") })
	})
	t.Run("InvalidPhoto", func(t *testing.T) {
		conf := config.NewMinimalTestConfig(t.TempDir())
		conf.Options().SidecarYaml = true
		conf.Options().DisableBackups = false
		// A photo without a name cannot resolve a file name and must not panic or write.
		photo := &entity.Photo{PhotoUID: rnd.GenerateUID(entity.PhotoUID)}
		assert.NotPanics(t, func() { NewMeta(conf).saveSidecarYaml(photo, "test") })
	})
}
