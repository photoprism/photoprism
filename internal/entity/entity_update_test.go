package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/rnd"
)

// missingPhotoID returns a photo ID that is not present in the current test database.
func missingPhotoID() uint {
	id := uint(10010001)

	for FindPhoto(Photo{ID: id}) != nil {
		id++
	}

	return id
}

func TestUpdate(t *testing.T) {
	t.Run("IDMissing", func(t *testing.T) {
		uid := rnd.GenerateUID(PhotoUID)
		m := &Photo{ID: 0, PhotoUID: uid, UpdatedAt: Now(), CreatedAt: Now(), PhotoTitle: "Foo"}
		updatedAt := m.UpdatedAt

		err := Update(m, "ID", "PhotoUID")

		if err == nil {
			t.Fatal("error expected")
		}

		assert.ErrorContains(t, err, "new record")
		assert.Equal(t, m.UpdatedAt.UTC(), updatedAt.UTC())
	})
	t.Run("UIDMissing", func(t *testing.T) {
		m := &Photo{ID: PhotoFixtures.Get("Photo01").ID, PhotoUID: "", UpdatedAt: Now(), CreatedAt: Now(), PhotoTitle: "Foo"}
		updatedAt := m.UpdatedAt

		err := Update(m, "ID", "PhotoUID")

		if err == nil {
			t.Fatal("error expected")
		}

		assert.ErrorContains(t, err, "record keys missing")
		assert.Equal(t, m.UpdatedAt.UTC(), updatedAt.UTC())
	})
	t.Run("NotUpdated", func(t *testing.T) {
		uid := rnd.GenerateUID(PhotoUID)
		m := &Photo{ID: missingPhotoID(), PhotoUID: uid, UpdatedAt: Now().Add(-time.Hour), CreatedAt: Now(), PhotoTitle: "Foo"}
		updatedAt := m.UpdatedAt

		err := Update(m, "ID", "PhotoUID")

		if err == nil {
			t.Fatal("error expected")
		}

		assert.ErrorContains(t, err, "record not found")
		assert.Greater(t, m.UpdatedAt.UTC(), updatedAt.UTC())
	})
	t.Run("Photo01", func(t *testing.T) {
		m := PhotoFixtures.Pointer("Photo01")
		updatedAt := m.UpdatedAt

		// Should be updated without any issues.
		if err := Update(m, "ID", "PhotoUID"); err != nil {
			assert.Greater(t, m.UpdatedAt.UTC(), updatedAt.UTC())
			t.Fatal(err)
			return
		} else {
			assert.Greater(t, m.UpdatedAt.UTC(), updatedAt.UTC())
			t.Logf("(1) UpdatedAt: %s -> %s", updatedAt.UTC(), m.UpdatedAt.UTC())
			t.Logf("(1) Successfully updated values")
		}

		// Tests that no error is returned on MySQL/MariaDB although
		// the number of affected rows is 0.
		if err := Update(m, "ID", "PhotoUID"); err != nil {
			assert.Greater(t, m.UpdatedAt.UTC(), updatedAt.UTC())
			t.Fatal(err)
			return
		} else {
			assert.Greater(t, m.UpdatedAt.UTC(), updatedAt.UTC())
			t.Logf("(2) UpdatedAt: %s -> %s", updatedAt.UTC(), m.UpdatedAt.UTC())
			t.Logf("(2) Successfully updated values")
		}
	})
	t.Run("NonExistentKeys", func(t *testing.T) {
		m := PhotoFixtures.Pointer("Photo01")
		m.ID = missingPhotoID()
		m.PhotoUID = rnd.GenerateUID(PhotoUID)
		updatedAt := m.UpdatedAt
		if err := Update(m, "ID", "PhotoUID"); err == nil {
			t.Errorf("expected error: %#v", m)
		} else {
			assert.ErrorContains(t, err, "record not found")
			assert.Greater(t, m.UpdatedAt.UTC(), updatedAt.UTC())
		}
	})
}

func TestIsNewRecord(t *testing.T) {
	t.Run("ZeroKey", func(t *testing.T) {
		assert.True(t, isNewRecord(Db(), &Photo{}))
		assert.True(t, isNewRecord(Db(), &Album{AlbumUID: rnd.GenerateUID(AlbumUID)}))
	})
	t.Run("Key", func(t *testing.T) {
		assert.False(t, isNewRecord(Db(), &Photo{ID: 1}))
		assert.False(t, isNewRecord(Db(), &Client{ClientUID: rnd.GenerateUID(ClientUID)}))
	})
	t.Run("CompositeKey", func(t *testing.T) {
		assert.False(t, isNewRecord(Db(), &PhotoLabel{PhotoID: 1, LabelID: 2}))
		assert.True(t, isNewRecord(Db(), &PhotoLabel{LabelID: 2}))
	})
	t.Run("NoModel", func(t *testing.T) {
		assert.True(t, isNewRecord(Db(), "photo"))
	})
}

func TestColumnValues(t *testing.T) {
	t.Run("Columns", func(t *testing.T) {
		values := columnValues(Db(), &Photo{}, Values{"PhotoTitle": "Title", "TakenAt": time.Time{}, "Details": &Details{}, "PhotoDescription": "Text"})
		assert.Equal(t, Values{"photo_title": "Title", "taken_at": time.Time{}}, values)
	})
	t.Run("NoModel", func(t *testing.T) {
		values := Values{"PhotoTitle": "Title"}
		assert.Equal(t, values, columnValues(Db(), "photo", values))
	})
}
