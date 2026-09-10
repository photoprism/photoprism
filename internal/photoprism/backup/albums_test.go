package backup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestAlbums(t *testing.T) {
	backupPath, err := filepath.Abs("./testdata/albums")

	if err != nil {
		t.Fatal(err)
	}

	if err = os.MkdirAll(backupPath, fs.ModeDir); err != nil {
		t.Fatal(err)
	}

	count, err := Albums(backupPath, true)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 36, count)

	count, err = Albums(backupPath, false)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 0, count)

	if err = os.RemoveAll(backupPath); err != nil {
		t.Fatal(err)
	}
}

func TestRestoreAlbums(t *testing.T) {
	t.Run("EmptyFolder", func(t *testing.T) {
		backupPath, err := filepath.Abs("./testdata/albums")

		if err != nil {
			t.Fatal(err)
		}

		if err = os.MkdirAll(backupPath, fs.ModeDir); err != nil {
			t.Fatal(err)
		}

		count, err := RestoreAlbums(backupPath, true)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 0, count)

		if err = os.RemoveAll(backupPath); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("NewDatabase", func(t *testing.T) {
		backupPath, err := filepath.Abs("./testdata/albums")

		if err != nil {
			t.Fatal(err)
		}

		if err = os.MkdirAll(backupPath, fs.ModeDir); err != nil {
			t.Fatal(err)
		}

		count, err := Albums(backupPath, true)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 36, count)

		// Wipe the database
		entity.Entities.Truncate(entity.Db())
		entity.CreateDefaultFixtures()

		if photocount, err := query.CountPhotos(); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, 0, photocount)
		}

		// Would be nice to test this on false, but, the options have metadata backup files disabled
		count, err = RestoreAlbums(backupPath, true)

		if err != nil {
			t.Fatal(err)
		}

		// 1 album is deleted
		assert.Equal(t, 35, count)

		photocount := 0
		if photocount, err = query.CountPhotos(); err != nil {
			t.Fatal(err)
		} else {
			assert.NotEqual(t, 0, photocount)
		}

		bigcount := int64(0)
		if err = entity.UnscopedDb().Model(&entity.Photo{}).Where("photo_type = ?", entity.MediaRestoring).Count(&bigcount).Error; err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, int64(photocount), bigcount)

		if err = os.RemoveAll(backupPath); err != nil {
			t.Fatal(err)
		}
	})

}
