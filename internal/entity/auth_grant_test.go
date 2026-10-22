package entity

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// newTestGrant creates a grant for alice and the specified client and deletes it after the test.
func newTestGrant(t *testing.T, clientID string) *Grant {
	t.Helper()

	m := &Grant{
		UserUID:     UserFixtures.Pointer("alice").UserUID,
		UserName:    "alice",
		ClientID:    clientID,
		ClientName:  "Test App",
		Scope:       "mcp",
		Resource:    "https://photos.example.com/api/v1/mcp",
		RefreshHash: rnd.Sha256([]byte(rnd.AuthToken())),
	}

	require.NoError(t, m.Create())
	t.Cleanup(func() { _ = m.Delete() })

	return m
}

// newTestGrantSession stores an access token session for the grant.
func newTestGrantSession(t *testing.T, g *Grant) *Session {
	t.Helper()

	sess := NewSession(3600, 0)
	sess.SetUser(UserFixtures.Pointer("alice"))
	sess.AuthProvider = authn.ProviderOAuthApp.String()
	sess.AuthMethod = authn.MethodPKCE.String()
	sess.AuthID = g.GrantUID
	sess.AuthScope = g.Scope
	sess.AuthAudience = g.Resource
	require.NoError(t, sess.Create())

	return sess
}

func TestGrant_TableName(t *testing.T) {
	assert.Equal(t, "auth_grants", Grant{}.TableName())
}

func TestGrant_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := newTestGrant(t, "csgrantcreate001")
		assert.True(t, rnd.IsUID(m.GrantUID, GrantUID))
		assert.False(t, m.UsedAt.IsZero())

		found := FindGrant(m.GrantUID)
		require.NotNil(t, found)
		assert.Equal(t, "csgrantcreate001", found.ClientID)
	})
	t.Run("Incomplete", func(t *testing.T) {
		assert.Error(t, (&Grant{UserUID: "uqxetse3cy5eo9z2"}).Create())
		var m *Grant
		assert.Error(t, m.Create())
	})
}

func TestGrant_Rotate(t *testing.T) {
	t.Run("Once", func(t *testing.T) {
		m := newTestGrant(t, "csgrantrotate001")
		other := *m
		oldHash := m.RefreshHash
		newHash := rnd.Sha256([]byte(rnd.AuthToken()))

		changed, err := m.Rotate(oldHash, newHash)
		require.NoError(t, err)
		assert.True(t, changed)
		assert.Equal(t, 1, m.RefreshGen)

		found := FindGrant(m.GrantUID)
		require.NotNil(t, found)
		assert.Equal(t, newHash, found.RefreshHash)
		assert.Equal(t, 1, found.RefreshGen)

		changed, err = other.Rotate(oldHash, rnd.Sha256([]byte(rnd.AuthToken())))
		require.NoError(t, err)
		assert.False(t, changed)
	})
	t.Run("Invalid", func(t *testing.T) {
		changed, err := (&Grant{}).Rotate("a", "b")
		assert.Error(t, err)
		assert.False(t, changed)
	})
}

func TestGrant_Idle(t *testing.T) {
	m := &Grant{UsedAt: time.Now().UTC().Add(-time.Hour)}
	assert.True(t, m.Idle(time.Now().UTC().Add(-time.Minute)))
	assert.False(t, m.Idle(time.Now().UTC().Add(-2*time.Hour)))
	var empty *Grant
	assert.True(t, empty.Idle(time.Now()))
}

func TestGrant_Delete(t *testing.T) {
	t.Run("WithSessions", func(t *testing.T) {
		m := newTestGrant(t, "csgrantdelete001")
		sess := newTestGrantSession(t, m)

		require.NoError(t, m.Delete())
		assert.Nil(t, FindGrant(m.GrantUID))

		_, err := FindAnySession(sess.ID)
		assert.Error(t, err)
	})
	t.Run("NoUID", func(t *testing.T) {
		assert.Error(t, (&Grant{}).Delete())
	})
}

func TestFindUserGrants(t *testing.T) {
	m := newTestGrant(t, "csgrantuser00001")
	grants, err := FindUserGrants(m.UserUID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(grants), 1)

	empty, err := FindUserGrants("")
	assert.NoError(t, err)
	assert.Empty(t, empty)
}

func TestFindUserClientGrant(t *testing.T) {
	m := newTestGrant(t, "csgrantuserclnt1")
	found := FindUserClientGrant(m.UserUID, "csgrantuserclnt1")
	require.NotNil(t, found)
	assert.Equal(t, m.GrantUID, found.GrantUID)
	assert.Nil(t, FindUserClientGrant(m.UserUID, "csunknown0000001"))
	assert.Nil(t, FindUserClientGrant("", "csgrantuserclnt1"))
}

func TestDeleteClientGrants(t *testing.T) {
	m := newTestGrant(t, "csgrantclient001")
	sess := newTestGrantSession(t, m)
	assert.Equal(t, 1, DeleteClientGrants("csgrantclient001"))
	assert.Nil(t, FindGrant(m.GrantUID))
	_, err := FindAnySession(sess.ID)
	assert.Error(t, err)
	assert.Equal(t, 0, DeleteClientGrants(""))
}

func TestDeleteUserGrants(t *testing.T) {
	m := &Grant{UserUID: "uqxetse3cy5eo9z9", ClientID: "csgrantdeluser01", RefreshHash: rnd.Sha256([]byte(rnd.AuthToken()))}
	require.NoError(t, m.Create())
	assert.Equal(t, 1, DeleteUserGrants("uqxetse3cy5eo9z9"))
	assert.Nil(t, FindGrant(m.GrantUID))
}

func TestDeleteIdleGrants(t *testing.T) {
	idle := newTestGrant(t, "csgrantidle00001")
	used := newTestGrant(t, "csgrantused00001")
	require.NoError(t, UnscopedDb().Model(&Grant{}).Where("grant_uid = ?", idle.GrantUID).UpdateColumn("used_at", time.Now().UTC().Add(-100*24*time.Hour)).Error)

	assert.GreaterOrEqual(t, DeleteIdleGrants(time.Now().UTC().Add(-90*24*time.Hour)), 1)
	assert.Nil(t, FindGrant(idle.GrantUID))
	assert.NotNil(t, FindGrant(used.GrantUID))
}

func TestDeleteGrantSessions(t *testing.T) {
	m := newTestGrant(t, "csgrantsess00001")
	sess := newTestGrantSession(t, m)
	require.NoError(t, DeleteGrantSessions(m.GrantUID))
	_, err := FindAnySession(sess.ID)
	assert.Error(t, err)
	assert.Error(t, DeleteGrantSessions(""))
}

func TestFindSession_Audience(t *testing.T) {
	m := newTestGrant(t, "csgrantaudience1")
	sess := newTestGrantSession(t, m)

	t.Run("FindSession", func(t *testing.T) {
		_, err := FindSession(sess.ID)
		assert.ErrorIs(t, err, ErrSessionNotFound)
	})
	t.Run("FindAudienceSession", func(t *testing.T) {
		found, err := FindAudienceSession(sess.ID, m.Resource)
		require.NoError(t, err)
		assert.Equal(t, sess.ID, found.ID)
		assert.True(t, found.IsOAuthApp())
		assert.True(t, found.HasAudience())

		_, err = FindAudienceSession(sess.ID, "https://other.example.com/api/v1/mcp")
		assert.ErrorIs(t, err, ErrSessionNotFound)
	})
	t.Run("FindAnySession", func(t *testing.T) {
		found, err := FindAnySession(sess.ID)
		require.NoError(t, err)
		assert.Equal(t, sess.ID, found.ID)
	})
	t.Run("RegularSession", func(t *testing.T) {
		regular := NewSession(3600, 0)
		regular.SetUser(UserFixtures.Pointer("alice"))
		require.NoError(t, regular.Create())
		t.Cleanup(func() { _ = regular.Delete() })

		found, err := FindAudienceSession(regular.ID, m.Resource)
		require.NoError(t, err)
		assert.False(t, found.HasAudience())
		assert.False(t, found.IsOAuthApp())

		_, err = FindSession(regular.ID)
		assert.NoError(t, err)
	})
}

func TestFindGrant(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := newTestGrant(t, "csgrantfind00002")
		found := FindGrant(m.GrantUID)
		require.NotNil(t, found)
		assert.Equal(t, m.ClientID, found.ClientID)
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.Nil(t, FindGrant("abc"))
		assert.Nil(t, FindGrant(rnd.GenerateUID(GrantUID)))
	})
}

func TestLookupGrant(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := newTestGrant(t, "csgrantlookup001")
		found, err := LookupGrant(m.GrantUID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, m.ClientID, found.ClientID)
	})
	t.Run("NotFound", func(t *testing.T) {
		found, err := LookupGrant(rnd.GenerateUID(GrantUID))
		assert.NoError(t, err)
		assert.Nil(t, found)
		found, err = LookupGrant("abc")
		assert.NoError(t, err)
		assert.Nil(t, found)
	})
	t.Run("ReadError", func(t *testing.T) {
		m := newTestGrant(t, "csgrantlookup002")
		name := "test:fail_grant_reads_" + rnd.Base36(8)
		require.NoError(t, Db().Callback().Query().Before("gorm:query").Register(name, func(db *gorm.DB) {
			if db.Statement.Table == (Grant{}).TableName() {
				_ = db.AddError(errors.New("grant lookup failed"))
			}
		}))
		t.Cleanup(func() { _ = Db().Callback().Query().Remove(name) })
		found, err := LookupGrant(m.GrantUID)
		assert.Error(t, err)
		assert.Nil(t, found)
		assert.Nil(t, FindGrant(m.GrantUID))
	})
}

func TestDeleteGrants(t *testing.T) {
	a := newTestGrant(t, "csgrantdelete002")
	b := newTestGrant(t, "csgrantdelete003")
	assert.Equal(t, 2, DeleteGrants(Grants{*a, *b}))
	assert.Nil(t, FindGrant(a.GrantUID))
	assert.Nil(t, FindGrant(b.GrantUID))
	assert.Equal(t, 0, DeleteGrants(Grants{{}}))
}

func TestGrant_UniqueUserClient(t *testing.T) {
	m := newTestGrant(t, "csgrantunique001")
	other := &Grant{UserUID: m.UserUID, ClientID: m.ClientID, RefreshHash: rnd.Sha256([]byte(rnd.AuthToken()))}
	assert.Error(t, other.Create())
}

func TestFindUserGrants_Order(t *testing.T) {
	older := newTestGrant(t, "csgrantorder0001")
	newer := newTestGrant(t, "csgrantorder0002")
	require.NoError(t, UnscopedDb().Model(&Grant{}).Where("grant_uid = ?", older.GrantUID).UpdateColumn("used_at", time.Now().UTC().Add(-time.Hour)).Error)
	other := &Grant{UserUID: "uqxetse3cy5eo9z8", ClientID: "csgrantorder0003", RefreshHash: rnd.Sha256([]byte(rnd.AuthToken()))}
	require.NoError(t, other.Create())
	t.Cleanup(func() { _ = other.Delete() })

	grants, err := FindUserGrants(older.UserUID)
	require.NoError(t, err)

	var uids []string

	for _, g := range grants {
		assert.Equal(t, older.UserUID, g.UserUID)
		uids = append(uids, g.GrantUID)
	}

	assert.Less(t, slices.Index(uids, newer.GrantUID), slices.Index(uids, older.GrantUID))
	assert.NotContains(t, uids, other.GrantUID)
}

func TestFindAudienceSession(t *testing.T) {
	m := newTestGrant(t, "csgrantaudience2")
	sess := newTestGrantSession(t, m)
	found, err := FindAudienceSession(sess.ID, m.Resource)
	require.NoError(t, err)
	assert.Equal(t, sess.ID, found.ID)
	_, err = FindAudienceSession(sess.ID, "")
	assert.ErrorIs(t, err, ErrSessionNotFound)
}

func TestFindAnySession(t *testing.T) {
	m := newTestGrant(t, "csgrantanysess01")
	sess := newTestGrantSession(t, m)
	found, err := FindAnySession(sess.ID)
	require.NoError(t, err)
	assert.True(t, found.HasAudience())
	_, err = FindAnySession("abc")
	assert.Error(t, err)
}

func TestSession_IsOAuthApp(t *testing.T) {
	assert.True(t, (&Session{AuthProvider: authn.ProviderOAuthApp.String()}).IsOAuthApp())
	assert.False(t, (&Session{AuthProvider: authn.ProviderApplication.String()}).IsOAuthApp())
	var s *Session
	assert.False(t, s.IsOAuthApp())
}

func TestSession_HasAudience(t *testing.T) {
	assert.True(t, (&Session{AuthAudience: "https://photos.example.com/api/v1/mcp"}).HasAudience())
	assert.False(t, (&Session{}).HasAudience())
	var s *Session
	assert.False(t, s.HasAudience())
}

func TestGrantRefreshTokenRevokeForm(t *testing.T) {
	f := form.OAuthRevokeToken{Token: rnd.RefreshToken(rnd.GenerateUID(GrantUID))}
	require.NoError(t, f.Validate())
	assert.Equal(t, form.RefreshToken, f.TokenTypeHint)
}

func TestFindClientGrants(t *testing.T) {
	older := newTestGrant(t, "csgrantlist00001")
	newer := &Grant{UserUID: UserFixtures.Pointer("bob").UserUID, ClientID: "csgrantlist00001", Scope: "mcp", RefreshHash: rnd.Sha256([]byte(rnd.AuthToken()))}
	require.NoError(t, newer.Create())
	t.Cleanup(func() { _ = newer.Delete() })
	newTestGrant(t, "csgrantlist00002")
	require.NoError(t, UnscopedDb().Model(&Grant{}).Where("grant_uid = ?", older.GrantUID).UpdateColumn("used_at", time.Now().UTC().Add(-time.Hour)).Error)

	t.Run("Success", func(t *testing.T) {
		found, err := FindClientGrants("csgrantlist00001")
		require.NoError(t, err)
		require.Len(t, found, 2)
		assert.Equal(t, newer.GrantUID, found[0].GrantUID)
		assert.Equal(t, older.GrantUID, found[1].GrantUID)
	})
	t.Run("None", func(t *testing.T) {
		found, err := FindClientGrants("csgrantlist00003")
		require.NoError(t, err)
		assert.Empty(t, found)
		found, err = FindClientGrants("")
		require.NoError(t, err)
		assert.Empty(t, found)
	})
}

func TestCountClientGrants(t *testing.T) {
	newTestGrant(t, "csgrantcount0001")
	g := &Grant{UserUID: UserFixtures.Pointer("bob").UserUID, ClientID: "csgrantcount0001", Scope: "mcp", RefreshHash: rnd.Sha256([]byte(rnd.AuthToken()))}
	require.NoError(t, g.Create())
	t.Cleanup(func() { _ = g.Delete() })
	newTestGrant(t, "csgrantcount0002")

	t.Run("Success", func(t *testing.T) {
		counts, err := CountClientGrants([]string{"csgrantcount0001", "csgrantcount0002", "csgrantcount0003"})
		require.NoError(t, err)
		assert.Equal(t, map[string]int{"csgrantcount0001": 2, "csgrantcount0002": 1}, counts)
	})
	t.Run("Empty", func(t *testing.T) {
		counts, err := CountClientGrants(nil)
		require.NoError(t, err)
		assert.Empty(t, counts)
	})
}

func TestSuspendClientGrants(t *testing.T) {
	prev := GrantSuspendedHandler
	t.Cleanup(func() { GrantSuspendedHandler = prev })

	var suspended []string
	GrantSuspendedHandler = func(uid string) { suspended = append(suspended, uid) }

	t.Run("Success", func(t *testing.T) {
		m := newTestGrant(t, "csgrantsusp00001")
		sess := newTestGrantSession(t, m)
		other := newTestGrant(t, "csgrantsusp00002")
		otherSess := newTestGrantSession(t, other)

		assert.Equal(t, 1, SuspendClientGrants("csgrantsusp00001"))
		assert.Equal(t, []string{m.GrantUID}, suspended)

		// The grant stays, its access token is deleted, and other clients are not affected.
		assert.NotNil(t, FindGrant(m.GrantUID))
		_, err := FindAnySession(sess.ID)
		assert.Error(t, err)
		_, err = FindAnySession(otherSess.ID)
		assert.NoError(t, err)
	})
	t.Run("None", func(t *testing.T) {
		suspended = nil
		assert.Equal(t, 0, SuspendClientGrants("csgrantsusp00003"))
		assert.Equal(t, 0, SuspendClientGrants(""))
		assert.Empty(t, suspended)
	})
}
