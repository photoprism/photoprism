package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestNewSession(t *testing.T) {
	t.Run("NoSessionData", func(t *testing.T) {
		m := NewSession(UnixDay, UnixHour*6)

		assert.True(t, rnd.IsSessionID(m.ID))
		assert.False(t, m.CreatedAt.IsZero())
		assert.False(t, m.UpdatedAt.IsZero())
		assert.False(t, m.ExpiresAt().IsZero())
		assert.NotEmpty(t, m.ID)
		assert.NotNil(t, m.Data())
		assert.Equal(t, 0, len(m.Data().Tokens))
	})
	t.Run("EmptySessionData", func(t *testing.T) {
		m := NewSession(UnixDay, UnixHour*6)
		m.SetData(NewSessionData())

		assert.True(t, rnd.IsSessionID(m.ID))
		assert.False(t, m.CreatedAt.IsZero())
		assert.False(t, m.UpdatedAt.IsZero())
		assert.False(t, m.ExpiresAt().IsZero())
		assert.NotEmpty(t, m.ID)
		assert.NotNil(t, m.Data())
		assert.Equal(t, 0, len(m.Data().Tokens))
	})
	t.Run("WithSessionData", func(t *testing.T) {
		data := NewSessionData()
		data.Tokens = []string{"foo", "bar"}
		m := NewSession(UnixDay, UnixHour*6)
		m.SetData(data)

		assert.True(t, rnd.IsSessionID(m.ID))
		assert.False(t, m.CreatedAt.IsZero())
		assert.False(t, m.UpdatedAt.IsZero())
		assert.False(t, m.ExpiresAt().IsZero())
		assert.NotEmpty(t, m.ID)
		assert.NotNil(t, m.Data())
		assert.Len(t, m.Data().Tokens, 2)
		assert.Equal(t, "foo", m.Data().Tokens[0])
		assert.Equal(t, "bar", m.Data().Tokens[1])
	})
}

func TestSession_SetData(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		m := NewSession(UnixDay, UnixHour*6)

		assert.NotNil(t, m)

		sess := m.SetData(nil)

		assert.NotNil(t, sess)
		assert.NotEmpty(t, sess.ID)
		assert.Equal(t, sess.ID, m.ID)
	})
}

func TestSession_TimedOut(t *testing.T) {
	t.Run("NewSession", func(t *testing.T) {
		m := NewSession(UnixDay, UnixHour)
		assert.False(t, m.TimeoutAt().IsZero())
		assert.Equal(t, m.ExpiresAt(), m.TimeoutAt())
		assert.False(t, m.TimedOut())
	})
	t.Run("NoExpiration", func(t *testing.T) {
		m := NewSession(0, UnixHour)
		t.Logf("Timeout: %s, Expiration: %s", m.TimeoutAt().String(), m.ExpiresAt())
		assert.True(t, m.TimeoutAt().IsZero())
		assert.Equal(t, m.ExpiresAt(), m.TimeoutAt())
		assert.False(t, m.TimedOut())
		assert.True(t, m.ExpiresAt().IsZero())
	})
	t.Run("NoTimeout", func(t *testing.T) {
		m := NewSession(UnixDay, 0)
		t.Logf("Timeout: %s, Expiration: %s", m.TimeoutAt().String(), m.ExpiresAt())
		assert.False(t, m.TimeoutAt().IsZero())
		assert.Equal(t, m.ExpiresAt(), m.TimeoutAt())
		assert.False(t, m.TimedOut())
		assert.False(t, m.ExpiresAt().IsZero())
	})
	t.Run("TimedOut", func(t *testing.T) {
		m := NewSession(UnixDay, UnixHour)
		utc := UnixTime()

		m.LastActive = utc - (UnixHour + 1)

		assert.False(t, m.TimeoutAt().IsZero())
		assert.True(t, m.TimedOut())
	})
	t.Run("NotTimedOut", func(t *testing.T) {
		m := NewSession(UnixDay, UnixHour)
		utc := UnixTime()

		m.LastActive = utc - (UnixHour - 10)

		assert.False(t, m.TimeoutAt().IsZero())
		assert.False(t, m.TimedOut())
	})
}

func TestSession_Expired(t *testing.T) {
	t.Run("NewSession", func(t *testing.T) {
		m := NewSession(UnixDay, UnixHour)
		t.Logf("Timeout: %s, Expiration: %s", m.TimeoutAt().String(), m.ExpiresAt())
		assert.False(t, m.ExpiresAt().IsZero())
		assert.False(t, m.Expired())
		assert.False(t, m.TimeoutAt().IsZero())
		assert.False(t, m.TimedOut())
	})
	t.Run("NoExpiration", func(t *testing.T) {
		m := NewSession(0, 0)
		t.Logf("Timeout: %s, Expiration: %s", m.TimeoutAt().String(), m.ExpiresAt())
		assert.True(t, m.ExpiresAt().IsZero())
		assert.False(t, m.Expired())
		assert.True(t, m.TimeoutAt().IsZero())
		assert.False(t, m.TimedOut())
	})
	t.Run("NoExpiration", func(t *testing.T) {
		m := NewSession(0, 0)
		t.Logf("Timeout: %s, Expiration: %s", m.TimeoutAt().String(), m.ExpiresAt())
		assert.True(t, m.ExpiresAt().IsZero())
		assert.False(t, m.Expired())
		assert.True(t, m.TimeoutAt().IsZero())
		assert.False(t, m.TimedOut())
	})
	t.Run("Expired", func(t *testing.T) {
		m := NewSession(UnixDay, UnixHour)
		t.Logf("Timeout: %s, Expiration: %s", m.TimeoutAt().String(), m.ExpiresAt())
		utc := UnixTime()

		m.SessExpires = utc - 10

		assert.False(t, m.ExpiresAt().IsZero())
		assert.True(t, m.Expired())
		assert.False(t, m.TimeoutAt().IsZero())
		assert.True(t, m.TimedOut())
		assert.Equal(t, m.ExpiresAt(), m.TimeoutAt())
	})
	t.Run("NotExpired", func(t *testing.T) {
		m := NewSession(UnixDay, UnixHour)
		utc := UnixTime()

		m.SessExpires = utc + 10

		assert.False(t, m.ExpiresAt().IsZero())
		assert.False(t, m.Expired())
		assert.False(t, m.TimeoutAt().IsZero())
		assert.False(t, m.TimedOut())
		assert.Equal(t, m.ExpiresAt(), m.TimeoutAt())
	})
}
