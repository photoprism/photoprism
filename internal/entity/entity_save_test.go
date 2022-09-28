package entity

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestSave(t *testing.T) {
	var r = rand.New(rand.NewSource(time.Now().UnixNano()))

	t.Run("HasCreatedUpdatedAt", func(t *testing.T) {
		id := 99999 + r.Intn(10000)
		m := Photo{ID: uint(id), PhotoUID: rnd.GenerateUID(PhotoUID), UpdatedAt: TimeStamp(), CreatedAt: TimeStamp()}

		if err := m.Save(); err != nil {
			t.Fatal(err)
			return
		}

		assert.NotNil(t, FindPhoto(m))
	})
	t.Run("HasCreatedAt", func(t *testing.T) {
		id := 99999 + r.Intn(10000)
		m := Photo{ID: uint(id), PhotoUID: rnd.GenerateUID(PhotoUID), CreatedAt: TimeStamp()}

		if err := m.Save(); err != nil {
			t.Fatal(err)
			return
		}

		assert.NotNil(t, FindPhoto(m))
	})
	t.Run("NoCreatedAt", func(t *testing.T) {
		id := 99999 + r.Intn(10000)
		m := Photo{ID: uint(id), PhotoUID: rnd.GenerateUID(PhotoUID), CreatedAt: TimeStamp()}

		if err := m.Save(); err != nil {
			t.Fatal(err)
			return
		}

		assert.NotNil(t, FindPhoto(m))
	})
}
