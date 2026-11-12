package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
)

// TestWsAlbumUpdates checks that shared readers receive only visible album identifiers.
func TestWsAlbumUpdates(t *testing.T) {
	_, _, _ = NewApiTest()
	visitor := &entity.Session{}
	visitor.SetUser(&entity.Visitor)
	link := entity.NewLink("as6sg6bxpogaaba7", false, false)
	require.NoError(t, link.Save())
	t.Cleanup(func() { _ = link.Delete() })
	visitor.SetData(&entity.SessionData{Tokens: []string{link.LinkToken}})
	fields := event.Data{"entities": []string{"as6sg6bxpogaaba7", "as6sg6bxpogaaba8"}, "Title": "Withheld"}
	data, ok := wsAlbumUpdates(visitor, fields)
	require.True(t, ok)
	assert.Equal(t, event.Data{"entities": []string{"as6sg6bxpogaaba7"}}, data)
	_, ok = wsAlbumUpdates(visitor, event.Data{"entities": []string{"as6sg6bxpogaaba8"}})
	assert.False(t, ok)
	_, ok = wsAlbumUpdates(nil, fields)
	assert.False(t, ok)
	_, ok = wsAlbumUpdates(visitor, event.Data{"entities": "invalid"})
	assert.False(t, ok)
	visitor.AuthScope = "settings"
	_, ok = wsAlbumUpdates(visitor, fields)
	assert.False(t, ok)
	visitor.AuthScope = ""
	user := entity.Visitor
	expired := time.Now().Add(-time.Hour)
	user.ExpiresAt = &expired
	visitor.SetUser(&user)
	_, ok = wsAlbumUpdates(visitor, fields)
	assert.False(t, ok)
}

// TestWsAlbumUpdateActivity checks that server notifications do not extend idle sessions.
func TestWsAlbumUpdateActivity(t *testing.T) {
	_, _, _ = NewApiTest()
	session := newWsTestSession(t, "gandalf")
	before := time.Now().Add(-time.Minute).Unix()
	cached, err := entity.FindSession(session.ID)
	require.NoError(t, err)
	cached.LastActive = before
	require.NoError(t, entity.Db().Model(session).UpdateColumn("last_active", before).Error)
	_, _, _ = wsOutgoing("albums.updated", session.ID, *session.GetUser(), wsClientRole{}, event.Data{"entities": []string{"as6sg6bxpogaaba8"}})
	var loaded entity.Session
	require.NoError(t, entity.Db().First(&loaded, "id = ?", session.ID).Error)
	assert.Equal(t, before, loaded.LastActive)
	assert.Equal(t, before, cached.LastActive)
}
