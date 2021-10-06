package query

import (
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/stretchr/testify/assert"
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
	if err := PurgeOrphanCameras(); err != nil {
		t.Fatal(err)
	}
}

func TestPurgeUnusedLenses(t *testing.T) {
	if err := PurgeOrphanLenses(); err != nil {
		t.Fatal(err)
	}
}
