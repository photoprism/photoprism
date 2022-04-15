package entity

import (
	"math/rand"
	"testing"
	"time"

	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestSave(t *testing.T) {
	var r = rand.New(rand.NewSource(time.Now().UnixNano()))

	t.Run("HasCreatedUpdatedAt", func(t *testing.T) {
		id := 99999 + r.Intn(10000)
		m := Photo{ID: uint(id), PhotoUID: rnd.GenerateUID('p'), UpdatedAt: TimeStamp(), CreatedAt: TimeStamp()}

		if err := m.Save(); err != nil {
			t.Fatal(err)
			return
		}

		if err := m.Find(); err != nil {
			t.Fatal(err)
			return
		}
	})
	t.Run("HasCreatedAt", func(t *testing.T) {
		id := 99999 + r.Intn(10000)
		m := Photo{ID: uint(id), PhotoUID: rnd.GenerateUID('p'), CreatedAt: TimeStamp()}

		if err := m.Save(); err != nil {
			t.Fatal(err)
			return
		}

		if err := m.Find(); err != nil {
			t.Fatal(err)
			return
		}
	})
	t.Run("NoCreatedAt", func(t *testing.T) {
		id := 99999 + r.Intn(10000)
		m := Photo{ID: uint(id), PhotoUID: rnd.GenerateUID('p'), CreatedAt: TimeStamp()}

		if err := m.Save(); err != nil {
			t.Fatal(err)
			return
		}

		if err := m.Find(); err != nil {
			t.Fatal(err)
			return
		}
	})
}
