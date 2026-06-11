package query

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

func TestSelectedPhotoUIDsForSession(t *testing.T) {
	const (
		normalUID  = "ps6sg6be2lvl0yh7" // not private, not archived, not shared with guests
		privateUID = "ps6sg6be2lvl0y13" // "Photo06", private
	)
	uids := []string{normalUID, privateUID}

	t.Run("AdminSeesAll", func(t *testing.T) {
		scoped, err := SelectedPhotoUIDsForSession(uids, aclSession("alice"))
		assert.NoError(t, err)
		assert.ElementsMatch(t, uids, scoped)
	})
	t.Run("GuestExcludesPrivateAndUnshared", func(t *testing.T) {
		scoped, err := SelectedPhotoUIDsForSession(uids, aclSession("guest"))
		assert.NoError(t, err)
		assert.NotContains(t, scoped, privateUID)
		assert.NotContains(t, scoped, normalUID)
	})
	t.Run("NilSessionUnchanged", func(t *testing.T) {
		scoped, err := SelectedPhotoUIDsForSession(uids, nil)
		assert.NoError(t, err)
		assert.ElementsMatch(t, uids, scoped)
	})
	t.Run("AdminShortCircuitSkipsQuery", func(t *testing.T) {
		// Full-access sessions return the input verbatim without a scope query, so even an unknown
		// UID passes through (existence is checked by the caller's own lookup).
		in := []string{normalUID, "ps000000000unknown"}
		scoped, err := SelectedPhotoUIDsForSession(in, aclSession("alice"))
		assert.NoError(t, err)
		assert.Equal(t, in, scoped)
	})
	t.Run("EmptyInput", func(t *testing.T) {
		scoped, err := SelectedPhotoUIDsForSession(nil, aclSession("guest"))
		assert.NoError(t, err)
		assert.Empty(t, scoped)
	})
}

func TestPhotoSelection(t *testing.T) {
	albums := form.Selection{Albums: []string{"as6sg6bxpogaaba9", "as6sg6bitoga0004", "as6sg6bxpogaaba8", "as6sg6bxpogaaba7"}}

	months := form.Selection{Albums: []string{"as6sg6bipogaabj9"}}

	folders := form.Selection{Albums: []string{"as6sg6bipogaaba1", "as6sg6bipogaabj8"}}

	states := form.Selection{Albums: []string{"as6sg6bipogaab11", "as6sg6bipotaab12", "asjv2cw2eikl3cb3"}}

	t.Run("NoItemsSelected", func(t *testing.T) {
		f := form.Selection{
			Photos: []string{},
		}

		r, err := SelectedPhotos(f)

		assert.Equal(t, "no items selected", err.Error())
		assert.Empty(t, r)
	})
	t.Run("PhotosSelected", func(t *testing.T) {
		f := form.Selection{
			Photos: []string{"ps6sg6be2lvl0yh7", "ps6sg6be2lvl0yh8"},
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

		assert.Equal(t, 9, len(r))
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

		assert.Equal(t, 4, len(r))
		assert.IsType(t, entity.Photos{}, r)
	})
}
