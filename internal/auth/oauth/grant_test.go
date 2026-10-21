package oauth

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/auth/tokens"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
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

// failGrantReads makes queries of the grants table fail until the test ends.
func failGrantReads(t *testing.T) {
	t.Helper()
	name := "test:fail_grant_reads_" + rnd.Base36(8)
	require.NoError(t, entity.Db().Callback().Query().Before("gorm:query").Register(name, func(db *gorm.DB) {
		if db.Statement.Table == (entity.Grant{}).TableName() {
			_ = db.AddError(errors.New("grant lookup failed"))
		}
	}))
	t.Cleanup(func() { _ = entity.Db().Callback().Query().Remove(name) })
}

// refreshGrantOf returns the grant of a refresh token found with FindRefreshGrant, or nil.
func refreshGrantOf(raw string) *entity.Grant {
	m, _ := FindRefreshGrant(raw)
	return m
}

func TestNewGrant(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		grant, refreshToken := newTestGrant(t, "csgrantnew000001")
		assert.Equal(t, "mcp", grant.Scope)
		assert.Equal(t, "https://photos.example.com/api/v1/mcp", grant.Resource)
		assert.Equal(t, HashSecret(refreshToken), grant.RefreshHash)
		assert.NotEqual(t, refreshToken, grant.RefreshHash)
		value, err := tokens.VerifyRefresh(refreshToken)
		assert.NoError(t, err)
		assert.True(t, strings.HasPrefix(value, grant.GrantUID+"."))
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

		refreshed, second, err := RefreshGrant(refreshGrantOf(first), first, "csgrantrefresh01")
		require.NoError(t, err)
		assert.Equal(t, grant.GrantUID, refreshed.GrantUID)
		assert.NotEqual(t, first, second)
		assert.Equal(t, 1, refreshed.RefreshGen)

		_, third, err := RefreshGrant(refreshGrantOf(second), second, "csgrantrefresh01")
		require.NoError(t, err)
		assert.NotEmpty(t, third)
	})
	t.Run("ReuseRevokesGrant", func(t *testing.T) {
		grant, first := newTestGrant(t, "csgrantreuse0001")
		sess := NewAccessToken(grant, entity.FindUserByName("alice"), authn.GrantAuthorizationCode)
		require.NoError(t, sess.Create())

		_, second, err := RefreshGrant(refreshGrantOf(first), first, "csgrantreuse0001")
		require.NoError(t, err)

		revoked, _, err := RefreshGrant(refreshGrantOf(first), first, "csgrantreuse0001")
		assert.ErrorIs(t, err, ErrRefreshTokenReused)
		assert.True(t, errors.Is(err, ErrInvalidGrant))
		require.NotNil(t, revoked)
		assert.Nil(t, entity.FindGrant(grant.GrantUID))

		_, err = entity.FindAnySession(sess.ID)
		assert.Error(t, err)

		// The current refresh token no longer works either.
		_, _, err = RefreshGrant(refreshGrantOf(second), second, "csgrantreuse0001")
		assert.ErrorIs(t, err, ErrInvalidGrant)
	})
	t.Run("OtherClient", func(t *testing.T) {
		grant, first := newTestGrant(t, "csgrantotherc001")
		_, _, err := RefreshGrant(refreshGrantOf(first), first, "csgrantotherc002")
		assert.ErrorIs(t, err, ErrInvalidGrant)
		assert.NotNil(t, entity.FindGrant(grant.GrantUID))
	})
	t.Run("Idle", func(t *testing.T) {
		grant, first := newTestGrant(t, "csgrantidle00001")
		require.NoError(t, entity.UnscopedDb().Model(&entity.Grant{}).Where("grant_uid = ?", grant.GrantUID).
			UpdateColumn("used_at", time.Now().UTC().Add(-GrantMaxIdle-time.Hour)).Error)
		_, _, err := RefreshGrant(refreshGrantOf(first), first, "csgrantidle00001")
		assert.ErrorIs(t, err, ErrInvalidGrant)
		assert.Nil(t, entity.FindGrant(grant.GrantUID))
	})
	t.Run("Invalid", func(t *testing.T) {
		_, _, err := RefreshGrant(nil, "abc", "csgrantrefresh01")
		assert.ErrorIs(t, err, ErrInvalidGrant)
		_, _, err = RefreshGrant(nil, tokens.SignRefresh("gs5gfen1bgxz7s9i.0123456789abcdef0123456789abcdef0123456789abcdef", time.Now().Add(time.Minute)), "csgrantrefresh01")
		assert.ErrorIs(t, err, ErrInvalidGrant)
	})
	t.Run("UnsignedRefused", func(t *testing.T) {
		grant, _ := newTestGrant(t, "csgrantunsigned1")
		unsigned := rnd.RefreshToken(grant.GrantUID)
		_, _, err := RefreshGrant(grant, unsigned, "csgrantunsigned1")
		assert.ErrorIs(t, err, ErrInvalidGrant)
		assert.NotErrorIs(t, err, ErrRefreshTokenReused)
		assert.NotNil(t, entity.FindGrant(grant.GrantUID))
	})
	t.Run("ForgedRefused", func(t *testing.T) {
		grant, refreshToken := newTestGrant(t, "csgrantforged001")
		_, _, err := RefreshGrant(grant, flipLast(refreshToken), "csgrantforged001")
		assert.ErrorIs(t, err, ErrInvalidGrant)
		assert.NotNil(t, entity.FindGrant(grant.GrantUID))
	})
	t.Run("OtherGrant", func(t *testing.T) {
		grant, _ := newTestGrant(t, "csgrantother0001")
		_, other := newTestGrant(t, "csgrantother0002")
		_, _, err := RefreshGrant(grant, other, "csgrantother0001")
		assert.ErrorIs(t, err, ErrInvalidGrant)
		assert.NotNil(t, entity.FindGrant(grant.GrantUID))
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
		_, _, err := RefreshGrant(refreshGrantOf(first), first, "csgrantrevoke002")
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
		assert.ErrorIs(t, err, ErrInvalidToken)
		assert.Nil(t, revoked)
	})
	t.Run("Unsigned", func(t *testing.T) {
		grant, _ := newTestGrant(t, "csgrantrevoke005")
		unsigned := rnd.RefreshToken(grant.GrantUID)
		require.NoError(t, entity.UnscopedDb().Model(&entity.Grant{}).Where("grant_uid = ?", grant.GrantUID).
			UpdateColumn("refresh_hash", HashSecret(unsigned)).Error)
		revoked, err := RevokeRefreshToken(unsigned, "")
		assert.ErrorIs(t, err, ErrInvalidToken)
		assert.Nil(t, revoked)
		assert.NotNil(t, entity.FindGrant(grant.GrantUID))
	})
	t.Run("GrantNotFound", func(t *testing.T) {
		grant, refreshToken := newTestGrant(t, "csgrantrevoke006")
		require.NoError(t, grant.Delete())
		revoked, err := RevokeRefreshToken(refreshToken, "")
		assert.ErrorIs(t, err, ErrGrantNotFound)
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
		found, err := FindRefreshGrant(refreshToken)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, grant.GrantUID, found.GrantUID)

		// Finding the grant does not consume the token.
		_, _, err = RefreshGrant(found, refreshToken, "csgrantfind00001")
		assert.NoError(t, err)
	})
	t.Run("Invalid", func(t *testing.T) {
		found, err := FindRefreshGrant("abc")
		assert.ErrorIs(t, err, ErrInvalidToken)
		assert.ErrorIs(t, err, ErrInvalidGrant)
		assert.Nil(t, found)
	})
	t.Run("Unsigned", func(t *testing.T) {
		grant, _ := newTestGrant(t, "csgrantfind00002")
		_, err := FindRefreshGrant(rnd.RefreshToken(grant.GrantUID))
		assert.ErrorIs(t, err, ErrInvalidToken)
	})
	t.Run("Forged", func(t *testing.T) {
		_, refreshToken := newTestGrant(t, "csgrantfind00003")
		_, err := FindRefreshGrant(flipLast(refreshToken))
		assert.ErrorIs(t, err, ErrInvalidToken)
	})
	t.Run("Expired", func(t *testing.T) {
		grant, refreshToken := newTestGrant(t, "csgrantfind00004")
		value, _, _, ok := tokens.SplitValue(refreshToken)
		require.True(t, ok)
		expired := tokens.SignRefresh(value, time.Now().Add(-time.Second))
		_, err := FindRefreshGrant(expired)
		assert.ErrorIs(t, err, ErrTokenExpired)
		assert.Contains(t, err.Error(), grant.GrantUID)
		_, _, err = RefreshGrant(refreshGrantOf(expired), expired, "csgrantfind00004")
		assert.ErrorIs(t, err, ErrInvalidGrant)
		assert.NotNil(t, entity.FindGrant(grant.GrantUID))

		// An expired value of the current token neither rotates nor revokes the grant.
		_, _, err = RefreshGrant(grant, expired, "csgrantfind00004")
		assert.ErrorIs(t, err, ErrTokenExpired)
		assert.NotErrorIs(t, err, ErrRefreshTokenReused)
		assert.NotNil(t, entity.FindGrant(grant.GrantUID))
	})
	t.Run("ReadError", func(t *testing.T) {
		_, refreshToken := newTestGrant(t, "csgrantfind00006")
		failGrantReads(t)
		found, err := FindRefreshGrant(refreshToken)
		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrInvalidGrant)
		assert.Nil(t, found)
	})
	t.Run("GrantNotFound", func(t *testing.T) {
		grant, refreshToken := newTestGrant(t, "csgrantfind00005")
		require.NoError(t, grant.Delete())
		found, err := FindRefreshGrant(refreshToken)
		assert.ErrorIs(t, err, ErrGrantNotFound)
		assert.NotErrorIs(t, err, ErrInvalidToken)
		assert.ErrorIs(t, err, ErrInvalidGrant)
		assert.Nil(t, found)
	})
}

func TestVerifyRefreshToken(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert.NoError(t, VerifyRefreshToken(tokens.SignRefresh(rnd.RefreshToken("gs5gfen1bgxz7s9i"), time.Now().Add(time.Minute))))
	})
	t.Run("Invalid", func(t *testing.T) {
		for _, raw := range []string{"", rnd.RefreshToken("gs5gfen1bgxz7s9i"), tokens.SignCode(rnd.RefreshToken("gs5gfen1bgxz7s9i"), time.Now().Add(time.Minute))} {
			assert.ErrorIs(t, VerifyRefreshToken(raw), ErrInvalidToken, raw)
		}
	})
	t.Run("Expired", func(t *testing.T) {
		assert.ErrorIs(t, VerifyRefreshToken(tokens.SignRefresh(rnd.RefreshToken("gs5gfen1bgxz7s9i"), time.Now().Add(-time.Minute))), ErrTokenExpired)
	})
}

func TestRotateGrant(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m, raw := newTestGrant(t, "cs5gfen1bgxz7s9i")

		next, err := rotateGrant(m, HashSecret(raw))
		require.NoError(t, err)
		assert.NotEqual(t, raw, next)
		grantUID, err := refreshGrantUID(next)
		assert.NoError(t, err)
		assert.Equal(t, m.GrantUID, grantUID)
		assert.Equal(t, HashSecret(next), entity.FindGrant(m.GrantUID).RefreshHash)
	})
	t.Run("RotatedMeanwhile", func(t *testing.T) {
		m, _ := newTestGrant(t, "cs5gfen1bgxz7s9i")

		_, err := rotateGrant(m, HashSecret("stale"))
		assert.ErrorIs(t, err, ErrRefreshTokenReused)
		assert.Nil(t, entity.FindGrant(m.GrantUID))
	})
}

func TestNewRefreshToken(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		raw, err := newRefreshToken("gs5gfen1bgxz7s9i")
		require.NoError(t, err)
		grantUID, err := refreshGrantUID(raw)
		assert.NoError(t, err)
		assert.Equal(t, "gs5gfen1bgxz7s9i", grantUID)
		_, expires, _, ok := tokens.SplitValue(raw)
		assert.True(t, ok)
		assert.InDelta(t, time.Now().Add(RefreshTokenTTL).Unix(), expires, 5)
		assert.Greater(t, RefreshTokenTTL, GrantMaxIdle)
	})
	t.Run("Unconfigured", func(t *testing.T) {
		orig := *tokens.OAuth
		t.Cleanup(func() { *tokens.OAuth = orig })
		tokens.OAuth.Key = nil
		raw, err := newRefreshToken("gs5gfen1bgxz7s9i")
		assert.ErrorIs(t, err, ErrSigningKey)
		assert.Empty(t, raw)
	})
}

func TestRefreshGrantUID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		grantUID, err := refreshGrantUID(tokens.SignRefresh(rnd.RefreshToken("gs5gfen1bgxz7s9i"), time.Now().Add(time.Minute)))
		assert.NoError(t, err)
		assert.Equal(t, "gs5gfen1bgxz7s9i", grantUID)
	})
	t.Run("Invalid", func(t *testing.T) {
		for _, raw := range []string{"", rnd.RefreshToken("gs5gfen1bgxz7s9i"), tokens.SignCode(rnd.RefreshToken("gs5gfen1bgxz7s9i"), time.Now().Add(time.Minute)), tokens.SignRefresh(rnd.AuthToken(), time.Now().Add(time.Minute)), tokens.SignRefresh(rnd.AuthToken(), time.Now().Add(-time.Minute))} {
			grantUID, err := refreshGrantUID(raw)
			assert.ErrorIs(t, err, ErrInvalidToken, raw)
			assert.Empty(t, grantUID, raw)
		}
	})
	t.Run("Expired", func(t *testing.T) {
		grantUID, err := refreshGrantUID(tokens.SignRefresh(rnd.RefreshToken("gs5gfen1bgxz7s9i"), time.Now().Add(-time.Second)))
		assert.ErrorIs(t, err, ErrTokenExpired)
		assert.Contains(t, err.Error(), "gs5gfen1bgxz7s9i")
		assert.Empty(t, grantUID)
	})
}
