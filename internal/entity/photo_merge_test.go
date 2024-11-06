package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPhoto_Stackable(t *testing.T) {
	t.Run("IsStackable", func(t *testing.T) {
		m := Photo{ID: 1, PhotoUID: "pr32t8j3feogit2t", PhotoName: "foo", PhotoStack: IsStackable, TakenAt: Now(), TakenAtLocal: time.Time{}, TakenSrc: SrcMeta, TimeZone: "Europe/Berlin"}
		assert.True(t, m.Stackable())
	})
	t.Run("IsStacked", func(t *testing.T) {
		m := Photo{ID: 1, PhotoUID: "pr32t8j3feogit2t", PhotoName: "foo", PhotoStack: IsStacked, TakenAt: Now(), TakenAtLocal: time.Time{}, TakenSrc: SrcMeta, TimeZone: "Europe/Berlin"}
		assert.True(t, m.Stackable())
	})
	t.Run("NoName", func(t *testing.T) {
		m := Photo{ID: 1, PhotoUID: "pr32t8j3feogit2t", PhotoName: "", TakenAt: time.Time{}, TakenAtLocal: Now(), TakenSrc: SrcMeta, TimeZone: "Europe/Berlin"}
		assert.False(t, m.Stackable())
	})
	t.Run("IsUnstacked", func(t *testing.T) {
		m := Photo{ID: 1, PhotoUID: "pr32t8j3feogit2t", PhotoName: "foo", PhotoStack: IsUnstacked, TakenAt: Now(), TakenAtLocal: time.Time{}, TakenSrc: SrcMeta, TimeZone: "Europe/Berlin"}
		assert.False(t, m.Stackable())
	})
	t.Run("NoID", func(t *testing.T) {
		m := Photo{ID: 0, PhotoUID: "pr32t8j3feogit2t", PhotoName: "foo", PhotoStack: IsStacked, TakenAt: Now(), TakenAtLocal: time.Time{}, TakenSrc: SrcMeta, TimeZone: "Europe/Berlin"}
		assert.False(t, m.Stackable())
	})
	t.Run("NoPhotoUID", func(t *testing.T) {
		m := Photo{ID: 1, PhotoUID: "", PhotoName: "foo", PhotoStack: IsStacked, TakenAt: Now(), TakenAtLocal: time.Time{}, TakenSrc: SrcMeta, TimeZone: "Europe/Berlin"}
		assert.False(t, m.Stackable())
	})
}

func TestPhoto_IdenticalIdentical(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		photo := PhotoFixtures.Get("Photo19")

		result, err := photo.Identical(true, true)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("result: %#v", result)
		assert.Equal(t, 1, len(result))
	})

	t.Run("unstacked photo", func(t *testing.T) {
		photo := &Photo{PhotoStack: IsUnstacked, PhotoName: "testName"}

		result, err := photo.Identical(true, true)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(result))
	})
	t.Run("Success", func(t *testing.T) {
		photo := PhotoFixtures.Get("Photo23")

		result, err := photo.Identical(true, true)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("result: %#v", result)
		assert.Equal(t, 2, len(result))
	})
	t.Run("Success", func(t *testing.T) {
		photo := PhotoFixtures.Get("Photo23")
		result, err := photo.Identical(true, false)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("result: %#v", result)
		assert.Equal(t, 2, len(result))
	})
}

func TestPhoto_Merge(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		photo := PhotoFixtures.Get("Photo23")
		original, merged, err := photo.Merge(true, false)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 1000023, int(original.ID))
		assert.Equal(t, 1000024, int(merged[0].ID))
	})
}
