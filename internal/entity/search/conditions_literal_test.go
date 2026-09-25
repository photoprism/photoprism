package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestPhotos_NameLiteral(t *testing.T) {
	base := "zzl" + rnd.Base36(5)
	match := scopeBasePathPhoto(t, base, base+"_a")
	scopeBasePathPhoto(t, base, base+"Xa")

	search := func(name string) (uids []string) {
		results, _, err := Photos(form.SearchPhotos{Name: name, Count: 100})
		require.NoError(t, err)

		for _, r := range results {
			uids = append(uids, r.PhotoUID)
		}

		return uids
	}

	assert.Equal(t, []string{match.PhotoUID}, search(base+"_a"))
	assert.Len(t, search(base+"*a"), 2)
}

func TestCameras_Literal(t *testing.T) {
	base := "Zzl" + rnd.Base36(5)

	for _, name := range []string{base + "_a", base + "Xa"} {
		m := entity.NewCamera("", name)
		require.NoError(t, m.Create())
		t.Cleanup(func() { _ = entity.UnscopedDb().Delete(m).Error })
	}

	results, err := Cameras(form.SearchCameras{Query: base + "_a", Count: 100})
	require.NoError(t, err)

	if assert.Len(t, results, 1) {
		assert.Contains(t, results[0].CameraName, base+"_a")
	}
}

func TestLenses_Literal(t *testing.T) {
	base := "Zzl" + rnd.Base36(5)

	for _, name := range []string{base + "_a", base + "Xa"} {
		m := entity.NewLens("", name)
		require.NoError(t, m.Create())
		t.Cleanup(func() { _ = entity.UnscopedDb().Delete(m).Error })
	}

	results, err := Lenses(form.SearchLenses{Query: base + "_a", Count: 100})
	require.NoError(t, err)

	if assert.Len(t, results, 1) {
		assert.Contains(t, results[0].LensName, base+"_a")
	}
}

// literalTestAlbums creates two manual albums titled base+"_a" and base+"Xa", each with one picture,
// and returns the picture in the first album.
func literalTestAlbums(t *testing.T, base string) *entity.Photo {
	t.Helper()

	var first *entity.Photo

	for i, title := range []string{base + "_a", base + "Xa"} {
		album := entity.NewAlbum(title, entity.AlbumManual)
		require.NoError(t, album.Create())
		photo := scopeBasePathPhoto(t, base, title)
		require.NoError(t, entity.NewPhotoAlbum(photo.PhotoUID, album.AlbumUID).Create())

		t.Cleanup(func() {
			_ = entity.UnscopedDb().Where("album_uid = ?", album.AlbumUID).Delete(&entity.PhotoAlbum{}).Error
			_ = entity.UnscopedDb().Delete(album).Error
			entity.FlushAlbumCache()
		})

		if i == 0 {
			first = photo
		}
	}

	return first
}

func TestPhotos_AlbumLiteral(t *testing.T) {
	base := "Zzl" + rnd.Base36(5)
	match := literalTestAlbums(t, base)

	results, _, err := Photos(form.SearchPhotos{Album: base + "_a", Count: 100})
	require.NoError(t, err)

	var uids []string

	for _, r := range results {
		uids = append(uids, r.PhotoUID)
	}

	assert.Equal(t, []string{match.PhotoUID}, uids)
}

func TestAlbums_QueryLiteral(t *testing.T) {
	base := "Zzl" + rnd.Base36(5)
	literalTestAlbums(t, base)

	results, err := Albums(form.SearchAlbums{Query: base + "_a", Type: entity.AlbumManual, Count: 100})
	require.NoError(t, err)

	if assert.Len(t, results, 1) {
		assert.Equal(t, base+"_a", results[0].AlbumTitle)
	}
}
