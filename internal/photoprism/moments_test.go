package photoprism

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

func TestMoments_Start(t *testing.T) {
	conf := config.TestConfig()

	m := NewMoments(conf)
	err := m.Start()

	if err != nil {
		t.Fatal(err)
	}
}

// albumState returns a comparable snapshot of every album, keyed by UID.
func albumState(t *testing.T) map[string]string {
	t.Helper()

	var albums entity.Albums

	require.NoError(t, entity.UnscopedDb().Find(&albums).Error)

	state := make(map[string]string, len(albums))

	for _, a := range albums {
		state[a.AlbumUID] = fmt.Sprintf("%s|%s|%s|%s|%04d-%02d-%02d",
			a.AlbumType, a.AlbumTitle, a.AlbumSlug, a.AlbumPath, a.AlbumYear, a.AlbumMonth, a.AlbumDay)
	}

	return state
}

// TestMoments_StartRepeated runs the worker twice over a library that has a picture in the
// originals root and a folder album without a path, and requires the second pass to be a
// no-op. Date stamping that matches albums by path rather than by key rewrites unrelated
// albums here, which then makes the month lookup miss and duplicate or rename them.
func TestMoments_StartRepeated(t *testing.T) {
	conf := config.TestConfig()

	// A folder album whose path was never persisted, as clearDuplicateFolderAlbumPaths leaves it.
	pathless := &entity.Album{
		AlbumUID: "as6sg6bxpogaabm1", AlbumType: entity.AlbumFolder,
		AlbumTitle: "Pathless Folder", AlbumSlug: "pathless-folder-moments", AlbumPath: "",
	}
	require.NoError(t, entity.UnscopedDb().Create(pathless).Error)

	// A picture in the originals root, which indexes with an empty photo path.
	rootPhoto := &entity.Photo{
		PhotoUID: "ps6sg6bxpogaabm1", PhotoName: "momentsroot", PhotoPath: "",
		TakenAt:      time.Date(2004, 3, 7, 12, 0, 0, 0, time.UTC),
		TakenAtLocal: time.Date(2004, 3, 7, 12, 0, 0, 0, time.UTC),
		TakenSrc:     entity.SrcMeta, PhotoQuality: 3,
	}
	require.NoError(t, entity.UnscopedDb().Create(rootPhoto).Error)

	t.Cleanup(func() {
		entity.UnscopedDb().Exec("DELETE FROM albums WHERE album_uid = ?", "as6sg6bxpogaabm1")
		entity.UnscopedDb().Exec("DELETE FROM photos WHERE photo_uid = ?", "ps6sg6bxpogaabm1")
	})

	m := NewMoments(conf)

	require.NoError(t, m.Start())
	first := albumState(t)

	require.NoError(t, m.Start())
	second := albumState(t)

	assert.Equal(t, len(first), len(second), "second pass must not add or remove albums")

	for uid, want := range first {
		got, ok := second[uid]
		assert.True(t, ok, "album %s disappeared on the second pass", uid)
		assert.Equal(t, want, got, "album %s changed on the second pass", uid)
	}

	// Every month album must own a distinct year and month.
	var months entity.Albums
	require.NoError(t, entity.UnscopedDb().Where("album_type = ?", entity.AlbumMonth).Find(&months).Error)

	seen := make(map[string]string, len(months))

	for _, a := range months {
		key := fmt.Sprintf("%04d-%02d", a.AlbumYear, a.AlbumMonth)
		prev, dup := seen[key]
		assert.False(t, dup, "month albums %s and %s share year/month %s", prev, a.AlbumUID, key)
		seen[key] = a.AlbumUID
	}
}
