package batch

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestPreparePhotoSaveRequestsResolvesItemTitles(t *testing.T) {
	albumTitle := fmt.Sprintf("Batch Album %d", time.Now().UnixNano())
	labelTitle := fmt.Sprintf("Batch Label %d", time.Now().UnixNano())

	values := &PhotosForm{
		Albums: Items{
			Action: ActionUpdate,
			Items: []Item{
				{Title: albumTitle, Action: ActionAdd},
				{Title: albumTitle, Action: ActionAdd},
				{Title: albumTitle, Action: ActionRemove},
			},
		},
		Labels: Items{
			Action: ActionUpdate,
			Items: []Item{
				{Title: labelTitle, Action: ActionAdd},
				{Title: labelTitle, Action: ActionRemove},
			},
		},
	}

	requests, _, _ := PreparePhotoSaveRequests(search.PhotoResults{}, nil, values)
	require.Len(t, requests, 0)

	albumUID := values.Albums.Items[0].Value
	labelUID := values.Labels.Items[0].Value
	require.True(t, rnd.IsUID(albumUID, entity.AlbumUID))
	require.True(t, rnd.IsUID(labelUID, entity.LabelUID))
	require.Equal(t, albumUID, values.Albums.Items[1].Value)
	require.Equal(t, "", values.Albums.Items[2].Value)
	require.Equal(t, "", values.Labels.Items[1].Value)

	t.Cleanup(func() {
		if albumUID != "" {
			if album := entity.FindAlbum(entity.Album{AlbumUID: albumUID}); album != nil {
				_ = album.Delete()
			}
		}
		if labelUID != "" {
			var label entity.Label
			if err := entity.Db().Where("label_uid = ?", labelUID).First(&label).Error; err == nil {
				_ = label.Delete()
			}
		}
	})
}

func TestEnsureAlbumUIDCreatesAndReuses(t *testing.T) {
	title := fmt.Sprintf("Resolver Album %d", time.Now().UnixNano())

	uid := ensureAlbumUID(title)
	require.True(t, rnd.IsUID(uid, entity.AlbumUID))
	assert.Equal(t, uid, ensureAlbumUID(title))

	if album := entity.FindAlbum(entity.Album{AlbumUID: uid}); album != nil {
		t.Cleanup(func() { _ = album.Delete() })
	}
}

func TestEnsureLabelUIDCreatesAndReuses(t *testing.T) {
	title := fmt.Sprintf("Resolver Label %d", time.Now().UnixNano())

	uid := ensureLabelUID(title)
	require.True(t, rnd.IsUID(uid, entity.LabelUID))
	assert.Equal(t, uid, ensureLabelUID(title))

	t.Cleanup(func() {
		var label entity.Label
		if err := entity.Db().Where("label_uid = ?", uid).First(&label).Error; err == nil {
			_ = label.Delete()
		}
	})
}
