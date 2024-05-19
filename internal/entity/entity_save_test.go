package entity

import (
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestSave(t *testing.T) {
	t.Run("HasCreatedUpdatedAt", func(t *testing.T) {
		id := 99999 + rand.IntN(10000)
		m := Photo{ID: uint(id), PhotoUID: rnd.GenerateUID(PhotoUID), UpdatedAt: TimeStamp(), CreatedAt: TimeStamp()}

		if err := m.Save(); err != nil {
			t.Fatal(err)
			return
		}

		assert.NotNil(t, FindPhoto(m))
	})
	t.Run("HasCreatedAt", func(t *testing.T) {
		id := 99999 + rand.IntN(10000)
		m := Photo{ID: uint(id), PhotoUID: rnd.GenerateUID(PhotoUID), CreatedAt: TimeStamp()}

		if err := m.Save(); err != nil {
			t.Fatal(err)
			return
		}

		assert.NotNil(t, FindPhoto(m))
	})
	t.Run("NoCreatedAt", func(t *testing.T) {
		id := 99999 + rand.IntN(10000)
		m := Photo{ID: uint(id), PhotoUID: rnd.GenerateUID(PhotoUID), CreatedAt: TimeStamp()}

		if err := m.Save(); err != nil {
			t.Fatal(err)
			return
		}

		assert.NotNil(t, FindPhoto(m))
	})
}
