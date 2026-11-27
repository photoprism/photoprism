package query

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
)

// TestNotifyAlbumCoverChanges checks UID-only updates, unchanged covers, and cache invalidation.
func TestNotifyAlbumCoverChanges(t *testing.T) {
	a := autoCoverAlbum(t, entity.AlbumManual)
	_, err := entity.CachedAlbumByUID(a.AlbumUID)
	require.NoError(t, err)
	sub := event.Subscribe("albums.updated")
	t.Cleanup(func() { event.Unsubscribe(sub) })
	notify := notifyAlbumCoverChanges(entity.AlbumManual)
	require.NoError(t, entity.UpdateAlbum(a.AlbumUID, entity.Values{"thumb": "new-cover"}))
	notify()
	select {
	case message := <-sub.Receiver:
		assert.Contains(t, message.Fields["entities"], a.AlbumUID)
		assert.Len(t, message.Fields, 1)
	case <-time.After(time.Second):
		t.Fatal("expected an album update")
	}
	loaded, err := entity.CachedAlbumByUID(a.AlbumUID)
	require.NoError(t, err)
	assert.Equal(t, "new-cover", loaded.Thumb)
	notifyAlbumCoverChanges(entity.AlbumManual)()
	select {
	case <-sub.Receiver:
		t.Fatal("unchanged covers must not publish an update")
	case <-time.After(20 * time.Millisecond):
	}
}
