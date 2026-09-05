package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/react"
)

func TestNewReaction(t *testing.T) {
	ValidateFixtures(t)
	t.Run("Rainbow", func(t *testing.T) {
		m := NewReaction(FileFixtures.Pointer("bridge.jpg").FileUID, UserFixtures.Pointer("bob").UserUID).React("🌈")

		assert.Equal(t, FileFixtures.Pointer("bridge.jpg").FileUID, m.UID)
		assert.Equal(t, UserFixtures.Pointer("bob").UserUID, m.UserUID)
		assert.Equal(t, "🌈", m.Reaction)
		assert.Equal(t, react.Rainbow, m.Emoji())
	})
}

func TestReaction_Save(t *testing.T) {
	ValidateFixtures(t)
	t.Run("Rainbow", func(t *testing.T) {
		m := NewReaction(FileFixtures.Pointer("bridge.jpg").FileUID, UserFixtures.Pointer("bob").UserUID).React(react.Rainbow)

		if err := m.Save(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(m).Error) })
	})
	t.Run("RainbowPlus1", func(t *testing.T) {
		m := NewReaction(FileFixtures.Pointer("bridge.jpg").FileUID, UserFixtures.Pointer("bob").UserUID).React(react.Rainbow)
		if n := FindReaction(m.UID, m.UserUID); n == nil {
			if err := m.Save(); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(m).Error) })
		}
		n := FindReaction(m.UID, m.UserUID)
		expected := n.Reacted + 1
		if assert.NotNil(t, n) {
			err := n.Save()
			if assert.Empty(t, err) {
				t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(n).Error) })
				n2 := FindReaction(m.UID, m.UserUID)
				assert.Equal(t, expected, n2.Reacted)
			}
		}
	})
}

func TestFindReaction(t *testing.T) {
	ValidateFixtures(t)
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

func TestReaction_Delete(t *testing.T) {
	ValidateFixtures(t)
	m := NewReaction(FileFixtures.Pointer("bridge.jpg").FileUID, UserFixtures.Pointer("bob").UserUID).React(react.Rainbow)
	if n := FindReaction(m.UID, m.UserUID); n == nil {
		if err := m.Save(); err != nil {
			t.Fatal(err)
		}
	}
	n := FindReaction(m.UID, m.UserUID)
	if assert.Empty(t, n.Delete()) {
		assert.Nil(t, FindReaction(m.UID, m.UserUID))
	}
}
