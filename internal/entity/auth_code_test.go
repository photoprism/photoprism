package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/rnd"
)

// newTestAuthCode creates an authorization request or code with a random hash.
func newTestAuthCode(t *testing.T, codeType string, expiresIn time.Duration) *AuthCode {
	t.Helper()

	m := &AuthCode{
		CodeHash:    rnd.Sha256([]byte(rnd.AuthToken())),
		CodeType:    codeType,
		ClientID:    "cs5gfen1bgxz7s9i",
		RedirectURI: "https://claude.ai/api/mcp/auth_callback",
		ExpiresAt:   time.Now().UTC().Add(expiresIn),
	}

	require.NoError(t, m.Create())

	return m
}

func TestAuthCode_TableName(t *testing.T) {
	assert.Equal(t, "auth_codes", AuthCode{}.TableName())
}

func TestAuthCode_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := newTestAuthCode(t, AuthCodeCode, time.Minute)
		assert.NotZero(t, m.ID)
	})
	t.Run("Nil", func(t *testing.T) {
		var m *AuthCode
		assert.Error(t, m.Create())
	})
	t.Run("NoHash", func(t *testing.T) {
		assert.Error(t, (&AuthCode{CodeType: AuthCodeCode}).Create())
	})
	t.Run("DuplicateHash", func(t *testing.T) {
		m := newTestAuthCode(t, AuthCodeCode, time.Minute)
		assert.Error(t, (&AuthCode{CodeHash: m.CodeHash, CodeType: AuthCodeCode}).Create())
	})
}

func TestAuthCode_Expired(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.False(t, (&AuthCode{ExpiresAt: time.Now().UTC().Add(time.Minute)}).Expired())
	})
	t.Run("Expired", func(t *testing.T) {
		assert.True(t, (&AuthCode{ExpiresAt: time.Now().UTC().Add(-time.Second)}).Expired())
	})
	t.Run("Nil", func(t *testing.T) {
		var m *AuthCode
		assert.True(t, m.Expired())
	})
}

func TestFindAuthCode(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := newTestAuthCode(t, AuthCodeCode, time.Minute)
		found, err := FindAuthCode(m.CodeHash, AuthCodeCode)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, m.ID, found.ID)
		assert.Equal(t, m.ClientID, found.ClientID)
	})
	t.Run("OtherType", func(t *testing.T) {
		m := newTestAuthCode(t, AuthCodeRequest, time.Minute)
		found, err := FindAuthCode(m.CodeHash, AuthCodeCode)
		assert.NoError(t, err)
		assert.Nil(t, found)
	})
	t.Run("Empty", func(t *testing.T) {
		found, err := FindAuthCode("", AuthCodeCode)
		assert.NoError(t, err)
		assert.Nil(t, found)
	})
}

func TestAuthCode_Consume(t *testing.T) {
	t.Run("Once", func(t *testing.T) {
		m := newTestAuthCode(t, AuthCodeCode, time.Minute)
		other := *m

		deleted, err := m.Consume()
		require.NoError(t, err)
		assert.True(t, deleted)

		deleted, err = other.Consume()
		require.NoError(t, err)
		assert.False(t, deleted)
	})
	t.Run("NoID", func(t *testing.T) {
		deleted, err := (&AuthCode{}).Consume()
		assert.Error(t, err)
		assert.False(t, deleted)
	})
}

func TestAuthCode_Replace(t *testing.T) {
	t.Run("Once", func(t *testing.T) {
		m := newTestAuthCode(t, AuthCodeRequest, time.Minute)
		other := *m
		codeHash := rnd.Sha256([]byte(rnd.AuthToken()))
		expires := time.Now().UTC().Add(30 * time.Second)

		changed, err := m.Replace(codeHash, AuthCodeCode, Values{"user_uid": "uqxetse3cy5eo9z2", "expires_at": expires})
		require.NoError(t, err)
		assert.True(t, changed)
		assert.Equal(t, codeHash, m.CodeHash)
		assert.Equal(t, AuthCodeCode, m.CodeType)

		found, err := FindAuthCode(codeHash, AuthCodeCode)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, "uqxetse3cy5eo9z2", found.UserUID)
		assert.WithinDuration(t, expires, found.ExpiresAt, time.Second)

		// The original request can no longer be found or completed.
		gone, err := FindAuthCode(other.CodeHash, AuthCodeRequest)
		assert.NoError(t, err)
		assert.Nil(t, gone)

		changed, err = other.Replace(rnd.Sha256([]byte(rnd.AuthToken())), AuthCodeCode, nil)
		require.NoError(t, err)
		assert.False(t, changed)
	})
	t.Run("Expired", func(t *testing.T) {
		m := newTestAuthCode(t, AuthCodeRequest, -time.Second)
		changed, err := m.Replace(rnd.Sha256([]byte(rnd.AuthToken())), AuthCodeCode, nil)
		require.NoError(t, err)
		assert.False(t, changed)
		assert.Equal(t, AuthCodeRequest, m.CodeType)
	})
	t.Run("InvalidArguments", func(t *testing.T) {
		m := newTestAuthCode(t, AuthCodeRequest, time.Minute)

		changed, err := m.Replace("", AuthCodeCode, nil)
		assert.Error(t, err)
		assert.False(t, changed)

		changed, err = (&AuthCode{}).Replace("abc", AuthCodeCode, nil)
		assert.Error(t, err)
		assert.False(t, changed)
	})
}

func TestDeleteExpiredAuthCodes(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expired := newTestAuthCode(t, AuthCodeCode, -time.Minute)
		valid := newTestAuthCode(t, AuthCodeRequest, time.Minute)

		n, err := DeleteExpiredAuthCodes()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, n, int64(1))

		found, err := FindAuthCode(expired.CodeHash, AuthCodeCode)
		assert.NoError(t, err)
		assert.Nil(t, found)

		found, err = FindAuthCode(valid.CodeHash, AuthCodeRequest)
		assert.NoError(t, err)
		assert.NotNil(t, found)
	})
}
