package query

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/stretchr/testify/assert"
)

func TestPhotoSelection(t *testing.T) {
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
}

func TestFileSelection(t *testing.T) {
	t.Run("no items selected", func(t *testing.T) {
		f := form.Selection{
			Photos: []string{},
		}

		r, err := SelectedFiles(f, FileSelectionAll())

		assert.Equal(t, "no items selected", err.Error())
		assert.Empty(t, r)
	})
	t.Run("files selected", func(t *testing.T) {
		f := form.Selection{
			Photos: []string{"pt9jtdre2lvl0yh7", "pt9jtdre2lvl0yh8"},
		}

		r, err := SelectedFiles(f, FileSelectionAll())

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 3, len(r))
		assert.IsType(t, entity.Files{}, r)
	})
}
