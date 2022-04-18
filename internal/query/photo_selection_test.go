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
	none := form.Selection{Photos: []string{}}

	one := form.Selection{Photos: []string{"pt9jtdre2lvl0yh8"}}

	two := form.Selection{Photos: []string{"pt9jtdre2lvl0yh7", "pt9jtdre2lvl0yh8"}}

	many := form.Selection{
		Files:  []string{"ft8es39w45bnlqdw"},
		Photos: []string{"pt9jtdre2lvl0y21", "pt9jtdre2lvl0y19", "pr2xu7myk7wrbk38", "pt9jtdre2lvl0yh7", "pt9jtdre2lvl0yh8"},
	}

	t.Run("EmptySelection", func(t *testing.T) {
		sel := DownloadSelection(true, false, true)
		if results, err := SelectedFiles(none, sel); err == nil {
			t.Fatal("error expected")
		} else {
			assert.Empty(t, results)
		}
	})
	t.Run("DownloadSelectionRawSidecarPrivate", func(t *testing.T) {
		sel := DownloadSelection(true, true, false)
		if results, err := SelectedFiles(one, sel); err != nil {
			t.Fatal(err)
		} else {
			assert.Len(t, results, 2)
		}
	})
	t.Run("DownloadSelectionRawOriginals", func(t *testing.T) {
		sel := DownloadSelection(true, false, true)
		if results, err := SelectedFiles(two, sel); err != nil {
			t.Fatal(err)
		} else {
			assert.Len(t, results, 2)
		}
	})
	t.Run("ShareSelectionOriginals", func(t *testing.T) {
		sel := ShareSelection(false)
		if results, err := SelectedFiles(many, sel); err != nil {
			t.Fatal(err)
		} else {
			assert.Len(t, results, 6)
		}
	})
	t.Run("ShareSelectionPrimary", func(t *testing.T) {
		sel := ShareSelection(true)
		if results, err := SelectedFiles(many, sel); err != nil {
			t.Fatal(err)
		} else {
			assert.Len(t, results, 4)
		}
	})
}
