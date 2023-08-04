package entity

import (
	"testing"
	"time"

	"github.com/photoprism/photoprism/pkg/authn"

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

func TestSession_Expires(t *testing.T) {
	t.Run("Set expiry date", func(t *testing.T) {
		m := NewSession(UnixDay, UnixHour)
		initialExpiryDate := m.SessExpires
		m.Expires(time.Date(2035, 01, 15, 12, 30, 0, 0, time.UTC))
		finalExpiryDate := m.SessExpires
		assert.Greater(t, finalExpiryDate, initialExpiryDate)

	})
	t.Run("Try to set zero date", func(t *testing.T) {
		m := NewSession(UnixDay, UnixHour)
		initialExpiryDate := m.SessExpires
		m.Expires(time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC))
		finalExpiryDate := m.SessExpires
		assert.Equal(t, finalExpiryDate, initialExpiryDate)
	})
}

func TestDeleteExpiredSessions(t *testing.T) {
	assert.Equal(t, 0, DeleteExpiredSessions())
	m := NewSession(UnixDay, UnixHour)
	m.Expires(time.Date(2000, 01, 15, 12, 30, 0, 0, time.UTC))
	m.Save()
	assert.Equal(t, 1, DeleteExpiredSessions())
}

func TestSessionStatusUnauthorized(t *testing.T) {
	m := SessionStatusUnauthorized()
	assert.Equal(t, 401, m.Status)
	assert.IsType(t, &Session{}, m)
}

func TestSessionStatusForbidden(t *testing.T) {
	m := SessionStatusForbidden()
	assert.Equal(t, 403, m.Status)
	assert.IsType(t, &Session{}, m)
}

func TestFindSessionByRefID(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.Nil(t, FindSessionByRefID(""))
	})
	t.Run("alice", func(t *testing.T) {
		m := FindSessionByRefID("sessxkkcabcd")
		assert.Equal(t, "alice", m.UserName)
		assert.IsType(t, &Session{}, m)
	})
}

func TestSession_RegenerateID(t *testing.T) {
	m := NewSession(UnixDay, UnixHour)
	initialID := m.ID
	m.RegenerateID()
	finalID := m.ID
	assert.NotEqual(t, initialID, finalID)
}

func TestSession_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := FindSessionByRefID("sessxkkcxxxx")
		assert.Empty(t, m)
		s := &Session{
			ID:          "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7xxx",
			UserName:    "charles",
			SessExpires: UnixDay * 3,
			SessTimeout: UnixTime() + UnixWeek,
			RefID:       "sessxkkcxxxx",
		}

		err := s.Create()

		if err != nil {
			t.Fatal(err)
		}

		m2 := FindSessionByRefID("sessxkkcxxxx")
		assert.Equal(t, "charles", m2.UserName)
	})
	t.Run("Invalid RefID", func(t *testing.T) {
		m, _ := FindSession("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7111")
		assert.Empty(t, m)
		s := &Session{
			ID:          "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7111",
			UserName:    "charles",
			SessExpires: UnixDay * 3,
			SessTimeout: UnixTime() + UnixWeek,
			RefID:       "123",
		}

		err := s.Create()

		if err != nil {
			t.Fatal(err)
		}

		m2, _ := FindSession("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7111")

		assert.NotEqual(t, "123", m2.RefID)
	})
	t.Run("ID already exists", func(t *testing.T) {
		s := &Session{
			ID:          "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0",
			UserName:    "charles",
			SessExpires: UnixDay * 3,
			SessTimeout: UnixTime() + UnixWeek,
			RefID:       "sessxkkcxxxx",
		}

		err := s.Create()
		assert.Error(t, err)
	})
}

func TestSession_Save(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := FindSessionByRefID("sessxkkcxxxy")
		assert.Empty(t, m)
		s := &Session{
			ID:          "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7xxy",
			UserName:    "chris",
			SessExpires: UnixDay * 3,
			SessTimeout: UnixTime() + UnixWeek,
			RefID:       "sessxkkcxxxy",
		}

		err := s.Save()

		if err != nil {
			t.Fatal(err)
		}

		m2 := FindSessionByRefID("sessxkkcxxxy")
		assert.Equal(t, "chris", m2.UserName)
	})
}

func TestSession_Updates(t *testing.T) {
	m := FindSessionByRefID("sessxkkcabcd")
	assert.Equal(t, "alice", m.UserName)

	m.Updates(Session{UserName: "anton"})

	assert.Equal(t, "anton", m.UserName)
}

func TestSession_User(t *testing.T) {
	t.Run("alice", func(t *testing.T) {
		m := FindSessionByRefID("sessxkkcabcd")
		assert.Equal(t, "uqxetse3cy5eo9z2", m.User().UserUID)
	})
	t.Run("empty", func(t *testing.T) {
		m := &Session{}
		assert.Equal(t, "", m.User().UserUID)
	})
}

func TestSession_RefreshUser(t *testing.T) {
	t.Run("bob", func(t *testing.T) {
		m := FindSessionByRefID("sessxkkcabce")

		assert.Equal(t, "bob", m.Username())

		m.UserName = "bobby"

		assert.Equal(t, "bobby", m.Username())

		assert.Equal(t, "bob", m.RefreshUser().UserName)

		assert.Equal(t, "bob", m.Username())
	})
	t.Run("empty", func(t *testing.T) {
		m := &Session{}
		assert.Equal(t, "", m.RefreshUser().UserUID)
	})
}

func TestSession_SetProvider(t *testing.T) {
	m := FindSessionByRefID("sessxkkcabce")
	assert.Equal(t, authn.ProviderDefault, m.Provider())
	m.SetProvider("")
	assert.Equal(t, authn.ProviderDefault, m.Provider())
	m.SetProvider(authn.ProviderLink)
	assert.Equal(t, authn.ProviderLink, m.Provider())
	m.SetProvider(authn.ProviderDefault)
	assert.Equal(t, authn.ProviderDefault, m.Provider())
}

func TestSession_ChangePassword(t *testing.T) {
	m := FindSessionByRefID("sessxkkcabce")
	assert.Empty(t, m.PreviewToken)
	assert.Empty(t, m.DownloadToken)

	err := m.ChangePassword("photoprism123")

	if err != nil {
		t.Fatal(err)
	}

	assert.NotEmpty(t, m.PreviewToken)
	assert.NotEmpty(t, m.DownloadToken)

	err2 := m.ChangePassword("Bobbob123!")

	if err2 != nil {
		t.Fatal(err2)
	}
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
