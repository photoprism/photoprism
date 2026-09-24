package query

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestPurgeOrphans(t *testing.T) {
	fileName := "hd89e5yhb8p9h.jpg"

	if err := entity.AddDuplicate(
		fileName,
		entity.RootOriginals,
		"2cad9168fa6acc5c5c2965ddf6ec465ca42fd811",
		661858,
		time.Date(2019, 3, 6, 2, 6, 51, 0, time.UTC).Unix(),
	); err != nil {
		t.Fatal(err)
	}

	if err := PurgeOrphans(); err != nil {
		t.Fatal(err)
	}
}

func TestPurgeOrphanFiles(t *testing.T) {
	files, err := OrphanFiles()

	if err != nil {
		t.Fatal(err)
	}

	assert.IsType(t, entity.Files{}, files)

	t.Logf("%d oprhan files: %#v", len(files), files)

	if count, err := PurgeOrphanFiles(); err != nil {
		t.Fatal(err)
	} else if l := len(files); l != count {
		t.Errorf("found and removed files must match: %d <> %d", l, count)
	} else {
		t.Logf("removed %d orphan files", count)
	}

	if result, err := OrphanFiles(); err != nil {
		t.Fatal(err)
	} else if len(result) != 0 {
		t.Errorf("there should be no more orphan files")
	}
}

func TestPurgeFileDuplicates(t *testing.T) {
	fileName := "hd89e5yhb8p9h.jpg"

	if err := entity.AddDuplicate(
		fileName,
		entity.RootOriginals,
		"2cad9168fa6acc5c5c2965ddf6ec465ca42fd811",
		661858,
		time.Date(2019, 3, 6, 2, 6, 51, 0, time.UTC).Unix(),
	); err != nil {
		t.Fatal(err)
	}

	d := &entity.Duplicate{FileName: fileName, FileRoot: entity.RootOriginals}

	if err := d.Find(); err != nil {
		t.Fatal(err)
	}

	err := PurgeOrphanDuplicates()

	assert.NoError(t, err)

	dp := &entity.Duplicate{FileName: fileName, FileRoot: entity.RootOriginals}

	if err := dp.Find(); err == nil {
		t.Fatalf("duplicate should be removed: %+v", dp)
	}
}

func TestPurgeUnusedCountries(t *testing.T) {
	if err := PurgeOrphanCountries(); err != nil {
		t.Fatal(err)
	}
}

func TestPurgeUnusedCameras(t *testing.T) {
	added, created, err := entity.AddCamera("Minolta", "X-700")
	assert.NoError(t, err)
	assert.True(t, created)
	orphan := entity.FirstOrCreateCamera(entity.NewCamera("Minolta", "XD-7"))
	assert.NotZero(t, orphan.ID)
	t.Cleanup(func() {
		entity.FlushCameraCache()
		assert.NoError(t, UnscopedDb().Delete(&entity.Camera{}, "id IN (?)", []uint{added.ID, orphan.ID}).Error)
	})

	if err = PurgeOrphanCameras(); err != nil {
		t.Fatal(err)
	}

	// Orphans are removed unless they have been added manually.
	var count int
	assert.NoError(t, UnscopedDb().Model(&entity.Camera{}).Where("id = ?", added.ID).Count(&count).Error)
	assert.Equal(t, 1, count)
	assert.NoError(t, UnscopedDb().Model(&entity.Camera{}).Where("id = ?", orphan.ID).Count(&count).Error)
	assert.Equal(t, 0, count)
}

func TestPurgeUnusedLenses(t *testing.T) {
	added, created, err := entity.AddLens("Helios", "44-2 58mm f/2")
	assert.NoError(t, err)
	assert.True(t, created)
	orphan := entity.FirstOrCreateLens(entity.NewLens("Helios", "44M-4 58mm f/2"))
	assert.NotZero(t, orphan.ID)
	t.Cleanup(func() {
		entity.FlushLensCache()
		assert.NoError(t, UnscopedDb().Delete(&entity.Lens{}, "id IN (?)", []uint{added.ID, orphan.ID}).Error)
	})

	if err = PurgeOrphanLenses(); err != nil {
		t.Fatal(err)
	}

	// Orphans are removed unless they have been added manually.
	var count int
	assert.NoError(t, UnscopedDb().Model(&entity.Lens{}).Where("id = ?", added.ID).Count(&count).Error)
	assert.Equal(t, 1, count)
	assert.NoError(t, UnscopedDb().Model(&entity.Lens{}).Where("id = ?", orphan.ID).Count(&count).Error)
	assert.Equal(t, 0, count)
}
