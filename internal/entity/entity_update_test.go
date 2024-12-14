package entity

import (
	"math/rand/v2"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestEntity_Update(t *testing.T) {
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
		id := 99999 + rand.IntN(10000)
		m := &Photo{ID: uint(id), PhotoUID: "", UpdatedAt: Now(), CreatedAt: Now(), PhotoTitle: "Foo"}
		updatedAt := m.UpdatedAt

		err := Update(m, "ID", "PhotoUID")

		if err == nil {
			t.Fatal("error expected")
		}

		assert.ErrorContains(t, err, "record keys missing")
		assert.Equal(t, m.UpdatedAt.UTC(), updatedAt.UTC())
	})
	t.Run("NotUpdated", func(t *testing.T) {
		id := 99999 + rand.IntN(10000)
		uid := rnd.GenerateUID(PhotoUID)
		m := &Photo{ID: uint(id), PhotoUID: uid, UpdatedAt: time.Now(), CreatedAt: Now(), PhotoTitle: "Foo"}
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
		camera := PhotoFixtures.Pointer("Photo01").Camera

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

		// Make sure that a valid Sub Struct wasn't removed
		assert.Equal(t, camera.ID, m.Camera.ID)
		assert.Equal(t, camera.CameraDescription, m.Camera.CameraDescription)
		assert.Equal(t, camera.CameraMake, m.Camera.CameraMake)
		assert.Equal(t, camera.CameraModel, m.Camera.CameraModel)
		assert.Equal(t, camera.CameraName, m.Camera.CameraName)
	})
	t.Run("NonExistentKeys", func(t *testing.T) {
		m := PhotoFixtures.Pointer("Photo01")
		m.ID = uint(10000000 + rand.IntN(10000))
		m.PhotoUID = rnd.GenerateUID(PhotoUID)
		updatedAt := m.UpdatedAt
		if err := Update(m, "ID", "PhotoUID"); err == nil {
			t.Errorf("expected error: %#v", m)
		} else {
			assert.ErrorContains(t, err, "record not found")
			assert.Greater(t, m.UpdatedAt.UTC(), updatedAt.UTC())
		}
	})

	// This test is checking that GormV1 behaviour of ignoring Sub Structs stays true
	t.Run("InconsistentCameraVSCameraID", func(t *testing.T) {
		//m := Photo{}
		m := NewUserPhoto(true, UserFixtures.Pointer("mary").UserUID)
		assert.NotNil(t, m.Camera)
		photoQuery := UnscopedDb().First(&m, "photo_uid = ?", PhotoFixtures.Pointer("Photo01").PhotoUID)
		assert.Nil(t, photoQuery.Error)
		if photoQuery.Error != nil {
			t.FailNow()
		}

		assert.Equal(t, PhotoFixtures.Pointer("Photo01").ID, m.ID)
		assert.Equal(t, PhotoFixtures.Pointer("Photo01").CameraID, m.CameraID)
		assert.NotNil(t, m.Camera)
		if m.Camera != nil {
			assert.Equal(t, &UnknownCamera, m.Camera)
		}

		if err := Update(&m, "ID", "PhotoUID"); err != nil {
			t.Errorf("unexpected error: %#v", err)
			t.FailNow()
		}
		assert.Equal(t, PhotoFixtures.Pointer("Photo01").ID, m.ID)
		assert.Equal(t, PhotoFixtures.Pointer("Photo01").CameraID, m.CameraID)
		assert.NotNil(t, m.Camera)
		if m.Camera != nil {
			assert.Equal(t, &UnknownCamera, m.Camera)
		}

		// Put things back together.
		m = PhotoFixtures.Get("Photo01")
		if res := Db().Save(&m); res.Error != nil {
			t.Errorf("unexpected error: %#v", res.Error)
			t.FailNow()
		}
		assert.Equal(t, PhotoFixtures.Pointer("Photo01").ID, m.ID)
		assert.Equal(t, PhotoFixtures.Pointer("Photo01").CameraID, m.CameraID)
		assert.NotNil(t, m.Camera)
		if m.Camera != nil {
			assert.Equal(t, PhotoFixtures.Pointer("Photo01").Camera, m.Camera)
		}

	})
}
