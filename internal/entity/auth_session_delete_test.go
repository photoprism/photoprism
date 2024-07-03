package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/time/unix"
)

func TestDeleteSession(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		id := rnd.SessionID("77be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1")
		m := &Session{ID: id, DownloadToken: "download123", PreviewToken: "preview123"}
		CacheSession(m, time.Hour)
		r, _ := sessionCache.Get(id)
		assert.NotEmpty(t, r)
		err := DeleteSession(m)
		if err != nil {
			t.Fatal(err)
		}
		r2, _ := sessionCache.Get(id)
		assert.Empty(t, r2)
	})
	t.Run("invalidID", func(t *testing.T) {
		m := &Session{ID: "123-invalid", DownloadToken: "download123", PreviewToken: "preview123"}
		CacheSession(m, time.Hour)

		err := DeleteSession(m)

		assert.Error(t, err)
	})
}

func TestDeleteChildSessions(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		s := NewSession(3600, 0)
		assert.Equal(t, 0, DeleteChildSessions(s))
	})
}

func TestDeleteClientSessions(t *testing.T) {
	// Test client UID.
	clientUID := "cs5gfen1bgx00000"

	// Create new test client.
	client := NewClient()
	client.ClientUID = clientUID

	// Make sure no sessions exist yet and test missing arguments.
	assert.Equal(t, 0, DeleteClientSessions(&Client{}, authn.MethodUndefined, -1))
	assert.Equal(t, 0, DeleteClientSessions(client, authn.MethodOAuth2, -1))
	assert.Equal(t, 0, DeleteClientSessions(client, authn.MethodOAuth2, 0))
	assert.Equal(t, 0, DeleteClientSessions(&Client{}, authn.MethodDefault, 0))

	// Create 10 test client sessions.
	for i := 0; i < 10; i++ {
		sess := NewSession(3600, 0)
		sess.SetClient(client)

		if err := sess.Save(); err != nil {
			t.Fatal(err)
		}
	}

	// Check if the expected number of sessions is deleted until none are left.
	assert.Equal(t, 0, DeleteClientSessions(client, authn.MethodOAuth2, -1))
	assert.Equal(t, 0, DeleteClientSessions(client, authn.MethodDefault, 1))
	assert.Equal(t, 9, DeleteClientSessions(client, authn.MethodOAuth2, 1))
	assert.Equal(t, 1, DeleteClientSessions(client, authn.MethodOAuth2, 0))
	assert.Equal(t, 0, DeleteClientSessions(client, authn.MethodOAuth2, 0))
	assert.Equal(t, 0, DeleteClientSessions(client, authn.MethodUndefined, 0))
}

func TestDeleteExpiredSessions(t *testing.T) {
	assert.Equal(t, 0, DeleteExpiredSessions())
	m := NewSession(unix.Day, unix.Hour)
	m.Expires(time.Date(2000, 01, 15, 12, 30, 0, 0, time.UTC))
	if err := m.Save(); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, DeleteExpiredSessions())
}

func TestDeleteFromSessionCache(t *testing.T) {
	id := rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1")
	sessionCache.Flush()
	bob := FindSessionByRefID("sessxkkcabce")
	CacheSession(bob, time.Hour)
	r, b := sessionCache.Get(id)
	assert.NotEmpty(t, r)
	assert.True(t, b)
	DeleteFromSessionCache("")
	r2, b2 := sessionCache.Get(id)
	assert.NotEmpty(t, r2)
	assert.True(t, b2)
	DeleteFromSessionCache(id)
	r3, b3 := sessionCache.Get(id)
	assert.Empty(t, r3)
	assert.False(t, b3)
}
