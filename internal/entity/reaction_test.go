package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/react"
)

func TestNewReaction(t *testing.T) {
	t.Run("Rainbow", func(t *testing.T) {
		m := NewReaction(FileFixtures.Pointer("bridge.jpg").FileUID, UserFixtures.Pointer("bob").UserUID).React("🌈")

		assert.Equal(t, FileFixtures.Pointer("bridge.jpg").FileUID, m.UID)
		assert.Equal(t, UserFixtures.Pointer("bob").UserUID, m.UserUID)
		assert.Equal(t, "🌈", m.Reaction)
		assert.Equal(t, react.Rainbow, m.Emoji())
	})
}

func TestReaction_Save(t *testing.T) {
	t.Run("Rainbow", func(t *testing.T) {
		m := NewReaction(FileFixtures.Pointer("bridge.jpg").FileUID, UserFixtures.Pointer("bob").UserUID).React(react.Rainbow)

		if err := m.Save(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestFindReaction(t *testing.T) {
	photoUID := PhotoFixtures.Get("Photo01").PhotoUID
	userUID := UserFixtures.Pointer("alice").UserUID

	t.Run("PhotoAliceLove", func(t *testing.T) {
		if m := FindReaction(photoUID, userUID); m == nil {
			t.Fatal("result must not be nil")
		} else {
			assert.Equal(t, react.Love, m.Emoji())
		}
	})
}
