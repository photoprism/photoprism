package query

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

func TestPhotoSelection(t *testing.T) {
	albums := form.Selection{Albums: []string{"at9lxuqxpogaaba9", "at6axuzitogaaiax", "at9lxuqxpogaaba8", "at9lxuqxpogaaba7"}}

	months := form.Selection{Albums: []string{"at1lxuqipogaabj9"}}

	folders := form.Selection{Albums: []string{"at1lxuqipogaaba1", "at1lxuqipogaabj8"}}

	states := form.Selection{Albums: []string{"at1lxuqipogaab11", "at1lxuqipotaab12", "at1lxuqipotaab19"}}

	t.Run("no items selected", func(t *testing.T) {
		f := form.Selection{
			Photos: []string{},
		}

		r, err := SelectedPhotos(f)

		assert.Equal(t, "no items selected", err.Error())
		assert.Empty(t, r)
	})
	t.Run("photos selected", func(t *testing.T) {
		f := form.Selection{
			Photos: []string{"pt9jtdre2lvl0yh7", "pt9jtdre2lvl0yh8"},
		}

		r, err := SelectedPhotos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 2, len(r))
		assert.IsType(t, entity.Photos{}, r)
	})
	t.Run("FindAlbums", func(t *testing.T) {
		r, err := SelectedPhotos(albums)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 6, len(r))
		assert.IsType(t, entity.Photos{}, r)
	})
	t.Run("FindMonths", func(t *testing.T) {
		r, err := SelectedPhotos(months)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 0, len(r))
		assert.IsType(t, entity.Photos{}, r)
	})
	t.Run("FindFolders", func(t *testing.T) {
		r, err := SelectedPhotos(folders)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 2, len(r))
		assert.IsType(t, entity.Photos{}, r)
	})
	t.Run("FindStates", func(t *testing.T) {
		r, err := SelectedPhotos(states)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 1, len(r))
		assert.IsType(t, entity.Photos{}, r)
	})
}
