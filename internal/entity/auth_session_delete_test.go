package entity

import (
	"sort"
	"strings"
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
		m := &Session{ID: id, PreviewToken: "preview123"}
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
	t.Run("ReleasesTokens", func(t *testing.T) {
		id := rnd.SessionID("11be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1")
		m := &Session{ID: id, PreviewToken: "release-pv-token"}
		CacheSession(m, time.Hour)
		assert.True(t, PreviewToken.HasValue("release-pv-token"))

		if err := DeleteSession(m); err != nil {
			t.Fatal(err)
		}

		assert.True(t, PreviewToken.MissingValue("release-pv-token"))
	})
	t.Run("KeepsTokenWhileOtherSessionActive", func(t *testing.T) {
		first := &Session{ID: rnd.SessionID("22be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1"),
			PreviewToken: "shared-pv-token"}
		CacheSession(first, time.Hour)

		second := &Session{ID: rnd.SessionID("33be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1"),
			PreviewToken: "shared-pv-token"}
		CacheSession(second, time.Hour)

		// Removing the first session keeps the shared tokens because the second still uses them.
		if err := DeleteSession(first); err != nil {
			t.Fatal(err)
		}
		assert.True(t, PreviewToken.HasValue("shared-pv-token"))

		// Removing the last session releases the tokens.
		if err := DeleteSession(second); err != nil {
			t.Fatal(err)
		}
		assert.True(t, PreviewToken.MissingValue("shared-pv-token"))
	})
	t.Run("InvalidId", func(t *testing.T) {
		m := &Session{ID: "123-invalid", PreviewToken: "preview123"}
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
	assert.Equal(t, 0, DeleteClientSessions(&Client{}, authn.MethodUndefined, -1, ""))
	assert.Equal(t, 0, DeleteClientSessions(client, authn.MethodOAuth2, -1, ""))
	assert.Equal(t, 0, DeleteClientSessions(client, authn.MethodOAuth2, 0, ""))
	assert.Equal(t, 0, DeleteClientSessions(&Client{}, authn.MethodDefault, 0, ""))

	// Create 10 test client sessions.
	for range 10 {
		sess := NewSession(3600, 0)
		sess.SetClient(client)

		if err := sess.Save(); err != nil {
			t.Fatal(err)
		}
	}

	// Check if the expected number of sessions is deleted until none are left.
	assert.Equal(t, 0, DeleteClientSessions(client, authn.MethodOAuth2, -1, ""))
	assert.Equal(t, 0, DeleteClientSessions(client, authn.MethodDefault, 1, ""))
	assert.Equal(t, 9, DeleteClientSessions(client, authn.MethodOAuth2, 1, ""))
	assert.Equal(t, 1, DeleteClientSessions(client, authn.MethodOAuth2, 0, ""))
	assert.Equal(t, 0, DeleteClientSessions(client, authn.MethodOAuth2, 0, ""))
	assert.Equal(t, 0, DeleteClientSessions(client, authn.MethodUndefined, 0, ""))
}

// newClientSessions creates n sessions for the client, all stamped with createdAt, and returns
// their IDs sorted ascending so a case can pick the one the ID tiebreak ranks last.
// Whatever survives the case is removed again, so the shared test database is left as found.
func newClientSessions(t *testing.T, client *Client, n int, createdAt time.Time) []string {
	ids := make([]string, 0, n)

	t.Cleanup(func() {
		if _, err := client.DeleteSessions(); err != nil {
			t.Error(err)
		}
	})

	for range n {
		sess := NewSession(3600, 0)
		sess.SetClient(client)

		if err := sess.Save(); err != nil {
			t.Fatal(err)
		}

		if err := sess.Updates(Values{"created_at": createdAt}); err != nil {
			t.Fatal(err)
		}

		ids = append(ids, sess.ID)
	}

	sort.Strings(ids)

	return ids
}

func TestDeleteClientSessionsOrder(t *testing.T) {
	sameSecond := time.Date(2026, 9, 11, 10, 0, 0, 0, time.UTC)
	t.Run("KeepsNewest", func(t *testing.T) {
		client := NewClient()
		client.ClientUID = "cs5gfen1bgx00001"

		older := newClientSessions(t, client, 2, sameSecond.Add(-time.Hour))
		newer := newClientSessions(t, client, 2, sameSecond)

		assert.Equal(t, 2, DeleteClientSessions(client, authn.MethodOAuth2, 2, ""))
		assertSessions(t, newer, older)
	})
	t.Run("BreaksTiesByID", func(t *testing.T) {
		client := NewClient()
		client.ClientUID = "cs5gfen1bgx00002"

		ids := newClientSessions(t, client, 3, sameSecond)

		assert.Equal(t, 2, DeleteClientSessions(client, authn.MethodOAuth2, 1, ""))
		assertSessions(t, ids[2:], ids[:2])
	})
	t.Run("ReservedRanksFirstInItsSecond", func(t *testing.T) {
		client := NewClient()
		client.ClientUID = "cs5gfen1bgx00003"

		// The lowest ID is the one the tiebreak alone would delete first.
		ids := newClientSessions(t, client, 4, sameSecond)

		assert.Equal(t, 3, DeleteClientSessions(client, authn.MethodOAuth2, 1, ids[0]))
		assertSessions(t, ids[:1], ids[1:])
	})
	t.Run("ReservedCountsTowardLimit", func(t *testing.T) {
		client := NewClient()
		client.ClientUID = "cs5gfen1bgx00004"

		ids := newClientSessions(t, client, 4, sameSecond)

		assert.Equal(t, 2, DeleteClientSessions(client, authn.MethodOAuth2, 2, ids[0]))
		assertSessions(t, []string{ids[0], ids[3]}, []string{ids[1], ids[2]})
	})
	t.Run("ReservedDoesNotOutrankANewerSession", func(t *testing.T) {
		client := NewClient()
		client.ClientUID = "cs5gfen1bgx00005"

		older := newClientSessions(t, client, 1, sameSecond.Add(-time.Hour))
		newer := newClientSessions(t, client, 1, sameSecond)

		assert.Equal(t, 1, DeleteClientSessions(client, authn.MethodOAuth2, 1, older[0]))
		assertSessions(t, newer, older)
	})
	t.Run("UnknownID", func(t *testing.T) {
		client := NewClient()
		client.ClientUID = "cs5gfen1bgx00006"

		ids := newClientSessions(t, client, 3, sameSecond)

		assert.Equal(t, 2, DeleteClientSessions(client, authn.MethodOAuth2, 1, "sessxkkcabcdefgh"))
		assertSessions(t, ids[2:], ids[:2])
	})
	t.Run("OtherClientID", func(t *testing.T) {
		other := NewClient()
		other.ClientUID = "cs5gfen1bgx00007"
		foreign := newClientSessions(t, other, 1, sameSecond)

		client := NewClient()
		client.ClientUID = "cs5gfen1bgx00008"
		ids := newClientSessions(t, client, 2, sameSecond)

		assert.Equal(t, 1, DeleteClientSessions(client, authn.MethodOAuth2, 1, foreign[0]))
		assertSessions(t, append(ids[1:], foreign...), ids[:1])
	})
	t.Run("ZeroLimitDeletesAll", func(t *testing.T) {
		client := NewClient()
		client.ClientUID = "cs5gfen1bgx00009"

		ids := newClientSessions(t, client, 3, sameSecond)

		assert.Equal(t, 3, DeleteClientSessions(client, authn.MethodOAuth2, 0, ids[2]))
		assertSessions(t, nil, ids)
	})
	t.Run("NilClient", func(t *testing.T) {
		assert.Equal(t, 0, DeleteClientSessions(nil, authn.MethodOAuth2, 1, ""))
	})
}

// assertSessions verifies that every ID in kept still resolves and every ID in gone does not.
func assertSessions(t *testing.T, kept, gone []string) {
	for _, id := range kept {
		if _, err := FindSession(id); err != nil {
			t.Errorf("session %s should have been retained: %s", id, err)
		}
	}

	for _, id := range gone {
		if _, err := FindSession(id); err == nil {
			t.Errorf("session %s should have been deleted", id)
		}
	}
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

func TestSessionSaveTokenLimit(t *testing.T) {
	// Session IDs that outrank any generated one, so the session saved below is retained
	// only because it is the one being saved.
	highIDs := []string{strings.Repeat("f", 64), strings.Repeat("e", 64)}

	// A synthetic user keeps these sessions out of the fixture accounts other tests count.
	user := &User{UserUID: rnd.GenerateUID(UserUID), UserName: "session-save-token-limit"}

	newAppSession := func(id string) *Session {
		sess := NewSession(3600, 0)

		if id != "" {
			sess.ID = id
		}

		sess.SetUser(user)
		sess.SetClientName("TestSessionSaveTokenLimit")
		sess.SetProvider(authn.ProviderApplication)
		sess.SetMethod(authn.MethodSession)

		return sess
	}

	// Stamp the existing sessions with the second the save will run in, since sessions
	// created in the same second are the case a save must not evict itself in.
	second := Now().Add(time.Second)

	for _, id := range highIDs {
		sess := newAppSession(id)

		if err := sess.Create(); err != nil {
			t.Fatal(err)
		} else if err = sess.Updates(Values{"created_at": second}); err != nil {
			t.Fatal(err)
		}
	}

	time.Sleep(time.Until(second))

	saved := newAppSession("")

	if err := saved.Save(); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := saved.Delete(); err != nil {
			t.Error(err)
		}
	})

	assertSessions(t, []string{saved.ID}, highIDs)
}
