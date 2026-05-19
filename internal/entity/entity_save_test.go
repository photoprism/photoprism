package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestSave(t *testing.T) {
	t.Run("HasCreatedUpdatedAt", func(t *testing.T) {
		id := missingPhotoID()
		m := Photo{ID: id, PhotoUID: rnd.GenerateUID(PhotoUID), UpdatedAt: Now(), CreatedAt: Now()}

		if err := m.Save(); err != nil {
			t.Fatal(err)
			return
		}

		assert.Equal(t, id, m.ID)
		assert.NotNil(t, FindPhoto(m))
	})
	t.Run("HasCreatedAt", func(t *testing.T) {
		m := Photo{PhotoUID: rnd.GenerateUID(PhotoUID), CreatedAt: Now()}

		if err := m.Save(); err != nil {
			t.Fatal(err)
			return
		}

		assert.NotNil(t, FindPhoto(m))
	})
	t.Run("NoCreatedAt", func(t *testing.T) {
		m := Photo{PhotoUID: rnd.GenerateUID(PhotoUID), CreatedAt: Now()}

		if err := m.Save(); err != nil {
			t.Fatal(err)
			return
		}

		assert.NotNil(t, FindPhoto(m))
	})
}
