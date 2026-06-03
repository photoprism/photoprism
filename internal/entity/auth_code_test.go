package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOAuthCode(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		raw, m, err := NewOAuthCode(OAuthCodeSpec{
			ClientUID:           "cs5gfen1bgxz7s9i",
			UserUID:             "us5gfen1bgxz7s9i",
			RedirectURI:         "photoprism://callback",
			Scope:               "openid profile",
			CodeChallenge:       "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			CodeChallengeMethod: "S256",
		})
		require.NoError(t, err)
		require.NotEmpty(t, raw)
		require.NotNil(t, m)
		assert.NotZero(t, m.ID)
		assert.Equal(t, HashOAuthCode(raw), m.CodeHash)
		assert.False(t, m.IsExpired())
		require.NoError(t, m.Delete())
	})
	t.Run("MissingClient", func(t *testing.T) {
		_, _, err := NewOAuthCode(OAuthCodeSpec{UserUID: "u", RedirectURI: "r"})
		assert.Error(t, err)
	})
	t.Run("MissingUser", func(t *testing.T) {
		_, _, err := NewOAuthCode(OAuthCodeSpec{ClientUID: "c", RedirectURI: "r"})
		assert.Error(t, err)
	})
	t.Run("MissingRedirectURI", func(t *testing.T) {
		_, _, err := NewOAuthCode(OAuthCodeSpec{ClientUID: "c", UserUID: "u"})
		assert.Error(t, err)
	})
}

func TestFindOAuthCode(t *testing.T) {
	t.Run("FoundThenSingleUse", func(t *testing.T) {
		raw, _, err := NewOAuthCode(OAuthCodeSpec{ClientUID: "cs5gfen1bgxz7s9i", UserUID: "us5gfen1bgxz7s9i", RedirectURI: "photoprism://cb"})
		require.NoError(t, err)
		found, err := FindOAuthCode(raw)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, "cs5gfen1bgxz7s9i", found.ClientUID)
		require.NoError(t, found.Delete())
		gone, err := FindOAuthCode(raw)
		require.NoError(t, err)
		assert.Nil(t, gone, "redeemed code must not be found again")
	})
	t.Run("Empty", func(t *testing.T) {
		m, err := FindOAuthCode("")
		require.NoError(t, err)
		assert.Nil(t, m)
	})
	t.Run("Unknown", func(t *testing.T) {
		m, err := FindOAuthCode("nonexistent-code-value")
		require.NoError(t, err)
		assert.Nil(t, m)
	})
}

func TestOAuthCodeConsume(t *testing.T) {
	t.Run("DeletesOnce", func(t *testing.T) {
		raw, m, err := NewOAuthCode(OAuthCodeSpec{ClientUID: "cs5gfen1bgxz7s9i", UserUID: "us5gfen1bgxz7s9i", RedirectURI: "photoprism://cb"})
		require.NoError(t, err)
		deleted, err := m.Consume()
		require.NoError(t, err)
		assert.True(t, deleted, "first consume must report the row as deleted")
		gone, err := FindOAuthCode(raw)
		require.NoError(t, err)
		assert.Nil(t, gone, "consumed code must not be found again")
	})
	t.Run("SecondConsumeReportsNotDeleted", func(t *testing.T) {
		_, m, err := NewOAuthCode(OAuthCodeSpec{ClientUID: "cs5gfen1bgxz7s9i", UserUID: "us5gfen1bgxz7s9i", RedirectURI: "photoprism://cb"})
		require.NoError(t, err)
		deleted, err := m.Consume()
		require.NoError(t, err)
		require.True(t, deleted)
		deleted, err = m.Consume()
		require.NoError(t, err)
		assert.False(t, deleted, "second consume of the same row must report not deleted")
	})
	t.Run("ZeroID", func(t *testing.T) {
		deleted, err := (&OAuthCode{}).Consume()
		assert.Error(t, err)
		assert.False(t, deleted)
	})
}

func TestHashOAuthCode(t *testing.T) {
	h := HashOAuthCode("abc")
	assert.Len(t, h, 64)
	assert.Equal(t, h, HashOAuthCode("abc"))
	assert.NotEqual(t, h, HashOAuthCode("abd"))
}

func TestOAuthCodeIsExpired(t *testing.T) {
	assert.True(t, (*OAuthCode)(nil).IsExpired())
	assert.True(t, (&OAuthCode{ExpiresAt: time.Now().UTC().Add(-time.Minute)}).IsExpired())
	assert.False(t, (&OAuthCode{ExpiresAt: time.Now().UTC().Add(time.Minute)}).IsExpired())
}

func TestPurgeExpiredOAuthCodes(t *testing.T) {
	raw, m, err := NewOAuthCode(OAuthCodeSpec{ClientUID: "cs5gfen1bgxz7s9i", UserUID: "us5gfen1bgxz7s9i", RedirectURI: "photoprism://cb"})
	require.NoError(t, err)
	m.ExpiresAt = time.Now().UTC().Add(-time.Hour)
	require.NoError(t, Db().Save(m).Error)
	n, err := PurgeExpiredOAuthCodes()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, n, int64(1))
	gone, err := FindOAuthCode(raw)
	require.NoError(t, err)
	assert.Nil(t, gone)
}
