package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPhotos_Photos(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {

		photo1 := PhotoFixtures.Get("Photo08")
		photo2 := PhotoFixtures.Get("Photo07")

		photos := Photos{&photo1, &photo2}

		r := photos.Photos()

		assert.Equal(t, 2, len(r))
	})
}

func TestPhotos_Archived(t *testing.T) {
	deletedAt := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	archived := &Photo{PhotoUID: "ps6sg6be2lvl0y01", DeletedAt: &deletedAt, PhotoQuality: 3}
	visible := &Photo{PhotoUID: "ps6sg6be2lvl0y02", PhotoQuality: 3}
	removed := &Photo{PhotoUID: "ps6sg6be2lvl0y03", DeletedAt: &deletedAt, PhotoQuality: -1}

	t.Run("Mixed", func(t *testing.T) {
		photos := Photos{visible, archived, nil, removed}
		assert.Equal(t, Photos{archived}, photos.Archived())
		assert.Len(t, photos, 4)
	})
	t.Run("NoneArchived", func(t *testing.T) {
		result := Photos{visible, removed}.Archived()
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Empty(t, Photos{}.Archived())
		assert.Empty(t, Photos(nil).Archived())
	})
}
