package entity

import (
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/face"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/stretchr/testify/assert"
)

func TestUpdate(t *testing.T) {
	t.Run("HasCreatedUpdatedAt", func(t *testing.T) {
		m := NewFace(rnd.PPID('j'), SrcAuto, face.RandomEmbeddings(1, face.RegularFace))
		id := m.ID

		m.CreatedAt = time.Now()
		m.UpdatedAt = time.Now()

		if err := m.Save(); err != nil {
			t.Fatal(err)
			return
		}

		found := FindFace(id)

		assert.NotNil(t, found)
		assert.Equal(t, id, found.ID)
		assert.Greater(t, time.Now(), m.UpdatedAt)
		assert.Equal(t, found.CreatedAt.UTC(), m.CreatedAt.UTC())
	})
	t.Run("HasCreatedAt", func(t *testing.T) {
		m := NewFace(rnd.PPID('j'), SrcAuto, face.RandomEmbeddings(1, face.RegularFace))
		id := m.ID

		m.CreatedAt = time.Now()

		if err := m.Save(); err != nil {
			t.Fatal(err)
			return
		}

		found := FindFace(id)
		assert.NotNil(t, found)
		assert.Equal(t, id, found.ID)
		assert.Greater(t, time.Now().UTC(), m.UpdatedAt.UTC())
		assert.Equal(t, found.CreatedAt.UTC(), m.CreatedAt.UTC())
	})
	t.Run("NoCreatedAt", func(t *testing.T) {
		m := NewFace(rnd.PPID('j'), SrcAuto, face.RandomEmbeddings(1, face.RegularFace))
		id := m.ID

		if err := m.Save(); err != nil {
			t.Fatal(err)
			return
		}

		found := FindFace(id)
		assert.NotNil(t, found)
		assert.Equal(t, id, found.ID)
		assert.Greater(t, time.Now(), m.UpdatedAt.UTC())
		assert.Equal(t, found.CreatedAt.UTC(), m.CreatedAt.UTC())
	})
}
