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

// Regression pin for the user-reported batch-edit homophone bug:
// when only `问` exists (slug `wen`) and the user adds `吻` (same pinyin
// slug `wen`), the batch resolver must create a NEW `吻` label instead
// of collapsing onto the existing `问`. The pre-fix code routed through
// FindLabel, whose slug fallback returned `问` and silently dropped the
// user's typed label.
func TestEnsureLabelUIDDoesNotCollapseHomophones(t *testing.T) {
	first := entity.FirstOrCreateLabel(entity.NewLabel("问", 0))
	require.NotNil(t, first)
	require.True(t, first.HasUID())

	t.Cleanup(func() {
		entity.FlushLabelCache()
		_ = entity.Db().Unscoped().Delete(first).Error
	})

	uid := ensureLabelUID("吻")
	require.True(t, rnd.IsUID(uid, entity.LabelUID))
	assert.NotEqual(t, first.LabelUID, uid)

	var resolved entity.Label
	require.NoError(t, entity.Db().Where("label_uid = ?", uid).First(&resolved).Error)
	assert.Equal(t, "吻", resolved.LabelName)

	t.Cleanup(func() {
		_ = entity.Db().Unscoped().Delete(&resolved).Error
	})
}
