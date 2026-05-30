package query

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

// aclSession builds an in-memory session for the named user fixture.
func aclSession(name string) *entity.Session {
	s := &entity.Session{}
	s.SetUser(entity.UserFixtures.Pointer(name))
	return s
}

func TestSelectedFilesForSession(t *testing.T) {
	// Include private pictures in the base selection so the session scope is what filters them.
	o := FileSelection{Private: true, MaxSize: 1024 * MiB}
	private := form.Selection{Photos: []string{"ps6sg6be2lvl0y13"}} // "Photo06", private

	t.Run("AdminSeesPrivate", func(t *testing.T) {
		files, err := SelectedFilesForSession(private, o, aclSession("alice"))
		assert.NoError(t, err)
		assert.NotEmpty(t, files)
	})
	t.Run("GuestDeniedPrivate", func(t *testing.T) {
		files, err := SelectedFilesForSession(private, o, aclSession("guest"))
		assert.NoError(t, err)
		assert.Empty(t, files)
	})
	t.Run("NilMatchesSelectedFiles", func(t *testing.T) {
		frm := form.Selection{Photos: []string{"ps6sg6be2lvl0yh7"}}
		base, err := SelectedFiles(frm, o)
		assert.NoError(t, err)
		scoped, err := SelectedFilesForSession(frm, o, nil)
		assert.NoError(t, err)
		assert.Equal(t, len(base), len(scoped))
	})
}

func TestFileSelection(t *testing.T) {
	none := form.Selection{Photos: []string{}}

	one := form.Selection{Photos: []string{"ps6sg6be2lvl0yh8"}}

	two := form.Selection{Photos: []string{"ps6sg6be2lvl0yh7", "ps6sg6be2lvl0yh8"}}

	albums := form.Selection{Albums: []string{"as6sg6bxpogaaba9", "as6sg6bitoga0004", "as6sg6bxpogaaba8", "as6sg6bxpogaaba7"}}

	months := form.Selection{Albums: []string{"as6sg6bipogaabj9"}}

	folders := form.Selection{Albums: []string{"as6sg6bipogaaba1", "as6sg6bipogaabj8"}}

	states := form.Selection{Albums: []string{"as6sg6bipogaab11", "as6sg6bipotaab12", "asjv2cw2eikl3cb3"}}

	many := form.Selection{
		Files:  []string{"fs6sg6bw45bnlqdw"},
		Photos: []string{"ps6sg6be2lvl0y21", "ps6sg6be2lvl0y19", "ps6sg6byk7wrbk38", "ps6sg6be2lvl0yh7", "ps6sg6be2lvl0yh8"},
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
			assert.Len(t, results, 4)
		}
	})
	t.Run("ShareSelectionPrimary", func(t *testing.T) {
		sel := ShareSelection(true)
		if results, err := SelectedFiles(many, sel); err != nil {
			t.Fatal(err)
		} else {
			assert.Len(t, results, 6)
		}
	})
	t.Run("ShareAlbums", func(t *testing.T) {
		sel := ShareSelection(true)
		if results, err := SelectedFiles(albums, sel); err != nil {
			t.Fatal(err)
		} else {
			assert.Len(t, results, 10)
		}
	})
	t.Run("ShareMonths", func(t *testing.T) {
		sel := ShareSelection(true)
		if results, err := SelectedFiles(months, sel); err != nil {
			t.Fatal(err)
		} else {
			assert.Len(t, results, 0)
		}
	})
	t.Run("ShareFoldersOriginals", func(t *testing.T) {
		sel := ShareSelection(true)
		if results, err := SelectedFiles(folders, sel); err != nil {
			t.Fatal(err)
		} else {
			assert.Len(t, results, 4)
		}
	})
	t.Run("ShareFolders", func(t *testing.T) {
		sel := ShareSelection(false)
		if results, err := SelectedFiles(folders, sel); err != nil {
			t.Fatal(err)
		} else {
			log.Debugf("ShareFolders Results: %#v", results)
			assert.Len(t, results, 4)
		}
	})
	t.Run("ShareStatesOriginals", func(t *testing.T) {
		sel := ShareSelection(true)
		if results, err := SelectedFiles(states, sel); err != nil {
			t.Fatal(err)
		} else {
			assert.Len(t, results, 5)
		}
	})
	t.Run("ShareStates", func(t *testing.T) {
		sel := ShareSelection(false)
		if results, err := SelectedFiles(states, sel); err != nil {
			t.Fatal(err)
		} else {
			log.Debugf("ShareStates Result: %#v", results[0])
			assert.Len(t, results, 5)
		}
	})
}
