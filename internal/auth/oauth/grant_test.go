package oauth

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/authn"
)

// newTestGrant redeems a new code for alice and the client and returns the grant with its refresh token.
func newTestGrant(t *testing.T, clientID string) (*entity.Grant, string) {
	t.Helper()

	alice := entity.FindUserByName("alice")
	require.NotNil(t, alice)

	spec := testCodeSpec()
	spec.ClientID = clientID
	spec.UserUID = alice.UserUID

	raw, err := IssueCode(spec)
	require.NoError(t, err)
	code, err := RedeemCode(raw, spec.ClientID, spec.RedirectURI, testVerifier)
	require.NoError(t, err)

	grant, refreshToken, err := NewGrant(code, alice, "Test App")
	require.NoError(t, err)
	t.Cleanup(func() { _ = grant.Delete() })

	return grant, refreshToken
}

func TestNewGrant(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		grant, refreshToken := newTestGrant(t, "csgrantnew000001")
		assert.Equal(t, "mcp", grant.Scope)
		assert.Equal(t, "https://photos.example.com/api/v1/mcp", grant.Resource)
		assert.Equal(t, HashSecret(refreshToken), grant.RefreshHash)
		assert.NotEqual(t, refreshToken, grant.RefreshHash)
	})
	t.Run("ReplacesPrevious", func(t *testing.T) {
		first, _ := newTestGrant(t, "csgrantreplace01")
		second, _ := newTestGrant(t, "csgrantreplace01")
		assert.NotEqual(t, first.GrantUID, second.GrantUID)
		assert.Nil(t, entity.FindGrant(first.GrantUID))
		assert.NotNil(t, entity.FindGrant(second.GrantUID))
	})
	t.Run("OtherUser", func(t *testing.T) {
		bob := entity.FindUserByName("bob")
		require.NotNil(t, bob)
		_, _, err := NewGrant(&entity.AuthCode{UserUID: "uqxetse3cy5eo9z2", ClientID: "cs5gfen1bgxz7s9i"}, bob, "Test App")
		assert.Error(t, err)
	})
}

func TestNewAccessToken(t *testing.T) {
	grant, _ := newTestGrant(t, "csgrantaccess001")
	alice := entity.FindUserByName("alice")
	sess := NewAccessToken(grant, alice, authn.GrantAuthorizationCode)

	assert.Equal(t, alice.UserUID, sess.UserUID)
	assert.Empty(t, sess.ClientUID)
	assert.Equal(t, "Test App", sess.ClientName)
	assert.Equal(t, grant.GrantUID, sess.AuthID)
	assert.Equal(t, "mcp", sess.AuthScope)
	assert.Equal(t, grant.Resource, sess.AuthAudience)
	assert.True(t, sess.IsOAuthApp())
	assert.False(t, sess.IsClient())
	assert.Equal(t, authn.GrantAuthorizationCode.String(), sess.GrantType)
	assert.WithinDuration(t, time.Now().Add(AccessTokenTTL), sess.ExpiresAt(), 5*time.Second)
}

func TestRefreshGrant(t *testing.T) {
	t.Run("Rotation", func(t *testing.T) {
		grant, first := newTestGrant(t, "csgrantrefresh01")

		refreshed, second, err := RefreshGrant(first, "csgrantrefresh01")
		require.NoError(t, err)
		assert.Equal(t, grant.GrantUID, refreshed.GrantUID)
		assert.NotEqual(t, first, second)
		assert.Equal(t, 1, refreshed.RefreshGen)

		_, third, err := RefreshGrant(second, "csgrantrefresh01")
		require.NoError(t, err)
		assert.NotEmpty(t, third)
	})
	t.Run("ReuseRevokesGrant", func(t *testing.T) {
		grant, first := newTestGrant(t, "csgrantreuse0001")
		sess := NewAccessToken(grant, entity.FindUserByName("alice"), authn.GrantAuthorizationCode)
		require.NoError(t, sess.Create())

		_, second, err := RefreshGrant(first, "csgrantreuse0001")
		require.NoError(t, err)

		revoked, _, err := RefreshGrant(first, "csgrantreuse0001")
		assert.ErrorIs(t, err, ErrRefreshTokenReused)
		assert.True(t, errors.Is(err, ErrInvalidGrant))
		require.NotNil(t, revoked)
		assert.Nil(t, entity.FindGrant(grant.GrantUID))

		_, err = entity.FindAnySession(sess.ID)
		assert.Error(t, err)

		// The current refresh token no longer works either.
		_, _, err = RefreshGrant(second, "csgrantreuse0001")
		assert.ErrorIs(t, err, ErrInvalidGrant)
	})
	t.Run("OtherClient", func(t *testing.T) {
		grant, first := newTestGrant(t, "csgrantotherc001")
		_, _, err := RefreshGrant(first, "csgrantotherc002")
		assert.ErrorIs(t, err, ErrInvalidGrant)
		assert.NotNil(t, entity.FindGrant(grant.GrantUID))
	})
	t.Run("Idle", func(t *testing.T) {
		grant, first := newTestGrant(t, "csgrantidle00001")
		require.NoError(t, entity.UnscopedDb().Model(&entity.Grant{}).Where("grant_uid = ?", grant.GrantUID).
			UpdateColumn("used_at", time.Now().UTC().Add(-GrantMaxIdle-time.Hour)).Error)
		_, _, err := RefreshGrant(first, "csgrantidle00001")
		assert.ErrorIs(t, err, ErrInvalidGrant)
		assert.Nil(t, entity.FindGrant(grant.GrantUID))
	})
	t.Run("Invalid", func(t *testing.T) {
		_, _, err := RefreshGrant("abc", "csgrantrefresh01")
		assert.ErrorIs(t, err, ErrInvalidGrant)
		_, _, err = RefreshGrant("gs5gfen1bgxz7s9i.0123456789abcdef0123456789abcdef0123456789abcdef", "csgrantrefresh01")
		assert.ErrorIs(t, err, ErrInvalidGrant)
	})
}

func TestRevokeRefreshToken(t *testing.T) {
	t.Run("Current", func(t *testing.T) {
		grant, refreshToken := newTestGrant(t, "csgrantrevoke001")
		revoked, err := RevokeRefreshToken(refreshToken, "")
		require.NoError(t, err)
		require.NotNil(t, revoked)
		assert.Nil(t, entity.FindGrant(grant.GrantUID))
	})
	t.Run("NotCurrent", func(t *testing.T) {
		grant, first := newTestGrant(t, "csgrantrevoke002")
		_, _, err := RefreshGrant(first, "csgrantrevoke002")
		require.NoError(t, err)
		revoked, err := RevokeRefreshToken(first, "")
		require.NoError(t, err)
		assert.Nil(t, revoked)
		assert.NotNil(t, entity.FindGrant(grant.GrantUID))
	})
	t.Run("OtherClient", func(t *testing.T) {
		grant, refreshToken := newTestGrant(t, "csgrantrevoke003")
		revoked, err := RevokeRefreshToken(refreshToken, "csgrantrevoke004")
		require.NoError(t, err)
		assert.Nil(t, revoked)
		assert.NotNil(t, entity.FindGrant(grant.GrantUID))
	})
	t.Run("Invalid", func(t *testing.T) {
		revoked, err := RevokeRefreshToken("abc", "")
		assert.NoError(t, err)
		assert.Nil(t, revoked)
	})
}

func TestPurgeIdleGrants(t *testing.T) {
	grant, _ := newTestGrant(t, "csgrantpurge0001")
	require.NoError(t, entity.UnscopedDb().Model(&entity.Grant{}).Where("grant_uid = ?", grant.GrantUID).
		UpdateColumn("used_at", time.Now().UTC().Add(-GrantMaxIdle-time.Hour)).Error)
	PurgeIdleGrants()
	assert.Nil(t, entity.FindGrant(grant.GrantUID))
}

func TestFindRefreshGrant(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		grant, refreshToken := newTestGrant(t, "csgrantfind00001")
		found := FindRefreshGrant(refreshToken)
		require.NotNil(t, found)
		assert.Equal(t, grant.GrantUID, found.GrantUID)

		// Finding the grant does not consume the token.
		_, _, err := RefreshGrant(refreshToken, "csgrantfind00001")
		assert.NoError(t, err)
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.Nil(t, FindRefreshGrant("abc"))
	})
}
