package entity

import (
	"math/rand/v2"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestSave(t *testing.T) {
	t.Run("HasCreatedUpdatedAt", func(t *testing.T) {
		id := 99999 + rand.IntN(10000)
		m := Photo{ID: uint(id), PhotoUID: rnd.GenerateUID(PhotoUID), UpdatedAt: Now(), CreatedAt: Now()}

		if err := m.Save(); err != nil {
			t.Fatal(err)
			return
		}

		assert.NotNil(t, FindPhoto(m))
	})
	t.Run("HasCreatedAt", func(t *testing.T) {
		id := 99999 + rand.IntN(10000)
		m := Photo{ID: uint(id), PhotoUID: rnd.GenerateUID(PhotoUID), CreatedAt: Now()}

		if err := m.Save(); err != nil {
			t.Fatal(err)
			return
		}

		assert.NotNil(t, FindPhoto(m))
	})
	t.Run("NoCreatedAt", func(t *testing.T) {
		id := 99999 + rand.IntN(10000)
		m := Photo{ID: uint(id), PhotoUID: rnd.GenerateUID(PhotoUID), CreatedAt: Now()}

		if err := m.Save(); err != nil {
			t.Fatal(err)
			return
		}

		assert.NotNil(t, FindPhoto(m))
	})

	t.Run("NewParentPhotoWithNewChildDetails", func(t *testing.T) {
		stmt := Db()
		labelotter := Label{LabelName: "otterz", LabelSlug: "otterz"}
		var deletedAt = gorm.DeletedAt{Time: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true}
		labelsnake := Label{LabelName: "snakez", LabelSlug: "snakez", DeletedAt: deletedAt}

		err := labelsnake.Save()
		if err != nil {
			assert.Nil(t, err)
			t.FailNow()
		}

		err = labelotter.Save()
		if err != nil {
			assert.Nil(t, err)
			t.FailNow()
		}

		details := &Details{Keywords: "cow, flower, snakez, otterz"}
		photo := Photo{ID: 934567, Details: details}

		// This was failing with a foreign key constraint violation
		err = photo.Save()
		assert.Nil(t, err)
		if err != nil {
			UnscopedDb().Delete(&labelotter)
			UnscopedDb().Delete(&labelsnake)
			t.FailNow()
		}

		photo2 := Photo{ID: 934567}
		res := stmt.Preload("Details").First(&photo2)
		if res.Error != nil {
			assert.Nil(t, res.Error)
			t.FailNow()
		}

		assert.NotNil(t, photo2.Details)
		if photo2.Details != nil {
			assert.Equal(t, details.Keywords, photo2.Details.Keywords)
		}
	})
}
