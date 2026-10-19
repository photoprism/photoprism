package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// newGrantTestUser creates a user with a grant and an access token for it, and deletes them after the test.
func newGrantTestUser(t *testing.T, name string) (*User, *Grant, *Session) {
	t.Helper()

	u := NewUser()
	u.UserName = name
	u.UserRole = "admin"
	u.CanLogin = true
	require.NoError(t, u.Create())
	t.Cleanup(func() { _ = u.Delete() })

	g := &Grant{UserUID: u.UserUID, UserName: name, ClientID: "csgrantuser" + name[len(name)-5:], Scope: "mcp",
		Resource: "https://photos.example.com/api/v1/mcp", RefreshHash: rnd.Sha256([]byte(rnd.AuthToken()))}
	require.NoError(t, g.Create())
	t.Cleanup(func() { _ = g.Delete() })

	sess := NewSession(3600, 0)
	sess.SetUser(u)
	sess.AuthProvider = authn.ProviderOAuthApp.String()
	sess.AuthMethod = authn.MethodPKCE.String()
	sess.AuthID = g.GrantUID
	sess.AuthScope = g.Scope
	sess.AuthAudience = g.Resource
	require.NoError(t, sess.Create())

	return u, g, sess
}

func TestUser_RevokeSessions_Grants(t *testing.T) {
	t.Run("DerivedSessions", func(t *testing.T) {
		u, g, sess := newGrantTestUser(t, "grant-derived")
		u.RevokeDerivedSessions(nil)
		assert.Nil(t, FindGrant(g.GrantUID))
		_, err := FindAnySession(sess.ID)
		assert.Error(t, err)
	})
	t.Run("AllSessions", func(t *testing.T) {
		u, g, _ := newGrantTestUser(t, "grant-allsess")
		u.RevokeSessions(nil, authn.RevokeAllSessions)
		assert.Nil(t, FindGrant(g.GrantUID))
	})
	t.Run("LoginSessions", func(t *testing.T) {
		u, g, _ := newGrantTestUser(t, "grant-loginss")
		u.RevokeSessions(nil, authn.RevokeLoginSessions)
		assert.NotNil(t, FindGrant(g.GrantUID))
	})
}

func TestUser_Delete_Grants(t *testing.T) {
	u, g, sess := newGrantTestUser(t, "grant-deleted")
	code := &AuthCode{CodeHash: rnd.Sha256([]byte(rnd.AuthToken())), CodeType: AuthCodeCode, UserUID: u.UserUID}
	require.NoError(t, code.Create())

	require.NoError(t, u.Delete())
	assert.Nil(t, FindGrant(g.GrantUID))
	_, err := FindAnySession(sess.ID)
	assert.Error(t, err)

	found, err := FindAuthCode(code.CodeHash, AuthCodeCode)
	assert.NoError(t, err)
	assert.Nil(t, found)
}

func TestClient_Purge_Grants(t *testing.T) {
	client, err := AddClient(form.Client{ClientName: "Purge App", AuthMethod: authn.MethodPKCE.String(), AuthScope: "mcp", RedirectURIs: []string{"https://claude.ai/cb"}})
	require.NoError(t, err)

	g := &Grant{UserUID: UserFixtures.Pointer("alice").UserUID, ClientID: client.ClientUID, Scope: "mcp", RefreshHash: rnd.Sha256([]byte(rnd.AuthToken()))}
	require.NoError(t, g.Create())
	t.Cleanup(func() { _ = g.Delete() })

	require.NoError(t, client.Purge())
	assert.Nil(t, FindGrant(g.GrantUID))
}

func TestClient_Delete_Grants(t *testing.T) {
	client, err := AddClient(form.Client{ClientName: "Grant App", AuthMethod: authn.MethodPKCE.String(), AuthScope: "mcp", RedirectURIs: []string{"https://claude.ai/cb"}})
	require.NoError(t, err)

	g := &Grant{UserUID: UserFixtures.Pointer("alice").UserUID, ClientID: client.ClientUID, Scope: "mcp", RefreshHash: rnd.Sha256([]byte(rnd.AuthToken()))}
	require.NoError(t, g.Create())

	require.NoError(t, client.Delete())
	assert.Nil(t, FindGrant(g.GrantUID))
}

func TestGrantDeletedHandler(t *testing.T) {
	prev := GrantDeletedHandler
	t.Cleanup(func() { GrantDeletedHandler = prev })

	var deleted string
	GrantDeletedHandler = func(uid string) { deleted = uid }

	_, g, _ := newGrantTestUser(t, "grant-handler")
	require.NoError(t, g.Delete())
	assert.Equal(t, g.GrantUID, deleted)
}

func TestUser_RevokeGrants(t *testing.T) {
	t.Run("GrantsAndCodes", func(t *testing.T) {
		u, g, _ := newGrantTestUser(t, "grant-revokeall")
		code := &AuthCode{CodeHash: rnd.Sha256([]byte(rnd.AuthToken())), CodeType: AuthCodeCode, UserUID: u.UserUID}
		require.NoError(t, code.Create())

		assert.Equal(t, 1, u.RevokeGrants())
		assert.Nil(t, FindGrant(g.GrantUID))
		found, err := FindAuthCode(code.CodeHash, AuthCodeCode)
		assert.NoError(t, err)
		assert.Nil(t, found)
	})
	t.Run("Nil", func(t *testing.T) {
		var u *User
		assert.Equal(t, 0, u.RevokeGrants())
	})
}

func TestUser_RevokeDerivedSessions_Codes(t *testing.T) {
	u, _, _ := newGrantTestUser(t, "grant-codesrev")
	code := &AuthCode{CodeHash: rnd.Sha256([]byte(rnd.AuthToken())), CodeType: AuthCodeCode, UserUID: u.UserUID}
	require.NoError(t, code.Create())

	u.RevokeDerivedSessions(nil)

	found, err := FindAuthCode(code.CodeHash, AuthCodeCode)
	assert.NoError(t, err)
	assert.Nil(t, found)
}

func TestUser_SetPassword_Grants(t *testing.T) {
	u, g, _ := newGrantTestUser(t, "grant-password")
	require.NoError(t, u.SetPassword("NewPassword123!"))
	assert.Nil(t, FindGrant(g.GrantUID))
}

func TestUser_DeletePassword_Grants(t *testing.T) {
	u, _, _ := newGrantTestUser(t, "grant-delpassw")
	require.NoError(t, u.SetPassword("NewPassword123!"))

	g := &Grant{UserUID: u.UserUID, ClientID: "csgrantdelpass01", Scope: "mcp", RefreshHash: rnd.Sha256([]byte(rnd.AuthToken()))}
	require.NoError(t, g.Create())

	require.NoError(t, u.DeletePassword())
	assert.Nil(t, FindGrant(g.GrantUID))
}
