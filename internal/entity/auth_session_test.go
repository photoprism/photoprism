package entity

import (
	"testing"
	"time"

	"github.com/photoprism/photoprism/pkg/header"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestNewSession(t *testing.T) {
	t.Run("NoSessionData", func(t *testing.T) {
		m := NewSession(UnixDay, UnixHour*6)

		assert.True(t, rnd.IsAuthToken(m.AuthToken()))
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

		assert.True(t, rnd.IsAuthToken(m.AuthToken()))
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

		assert.True(t, rnd.IsAuthToken(m.AuthToken()))
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

func TestDeleteClientSessions(t *testing.T) {
	clientUID := "cs5gfen1bgx00000"

	// Make sure no sessions exist yet and test missing arguments.
	assert.Equal(t, 0, DeleteClientSessions("", -1))
	assert.Equal(t, 0, DeleteClientSessions(clientUID, -1))
	assert.Equal(t, 0, DeleteClientSessions(clientUID, 0))
	assert.Equal(t, 0, DeleteClientSessions("", 0))

	// Create 10 client sessions.
	for i := 0; i < 10; i++ {
		sess := NewSession(3600, 0)
		sess.SetClientIP(UnknownIP)
		sess.AuthID = clientUID
		sess.AuthProvider = authn.ProviderClient.String()
		sess.AuthMethod = authn.MethodOAuth2.String()
		sess.AuthScope = "*"

		if err := sess.Save(); err != nil {
			t.Fatal(err)
		}
	}

	// Check if the expected number of sessions is deleted until none are left.
	assert.Equal(t, 0, DeleteClientSessions(clientUID, -1))
	assert.Equal(t, 9, DeleteClientSessions(clientUID, 1))
	assert.Equal(t, 1, DeleteClientSessions(clientUID, 0))
	assert.Equal(t, 0, DeleteClientSessions(clientUID, 0))
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

func TestSession_Regenerate(t *testing.T) {
	t.Run("NewSession", func(t *testing.T) {
		m := NewSession(UnixDay, UnixHour)
		initialID := m.ID
		m.Regenerate()
		finalID := m.ID
		assert.NotEqual(t, initialID, finalID)
	})
	t.Run("Empty", func(t *testing.T) {
		m := Session{ID: ""}
		initialID := m.ID
		m.Regenerate()
		finalID := m.ID
		assert.NotEqual(t, initialID, finalID)
	})
	t.Run("Existing", func(t *testing.T) {
		m := Session{ID: "1234567"}
		initialID := m.ID
		m.Regenerate()
		finalID := m.ID
		assert.NotEqual(t, initialID, finalID)
	})
}

func TestSession_AuthToken(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		alice := SessionFixtures.Get("alice")
		sess := &Session{}
		assert.Equal(t, "", sess.ID)
		assert.Equal(t, "", sess.AuthToken())
		assert.False(t, rnd.IsSessionID(sess.ID))
		assert.False(t, rnd.IsAuthToken(sess.AuthToken()))
		assert.Equal(t, header.BearerAuth, sess.AuthTokenType())
		sess.Regenerate()
		assert.True(t, rnd.IsSessionID(sess.ID))
		assert.True(t, rnd.IsAuthToken(sess.AuthToken()))
		assert.Equal(t, header.BearerAuth, sess.AuthTokenType())
		sess.SetAuthToken(alice.AuthToken())
		assert.Equal(t, "a3859489780243a78b331bd44f58255b552dee104041a45c0e79b610f63af2e5", sess.ID)
		assert.Equal(t, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0", sess.AuthToken())
		assert.Equal(t, header.BearerAuth, sess.AuthTokenType())
	})
	t.Run("Alice", func(t *testing.T) {
		sess := SessionFixtures.Get("alice")
		assert.Equal(t, "a3859489780243a78b331bd44f58255b552dee104041a45c0e79b610f63af2e5", sess.ID)
		assert.Equal(t, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0", sess.AuthToken())
		assert.Equal(t, header.BearerAuth, sess.AuthTokenType())
	})
	t.Run("Find", func(t *testing.T) {
		alice := SessionFixtures.Get("alice")
		sess := FindSessionByRefID("sessxkkcabcd")
		assert.Equal(t, "a3859489780243a78b331bd44f58255b552dee104041a45c0e79b610f63af2e5", sess.ID)
		assert.Equal(t, "", sess.AuthToken())
		assert.Equal(t, header.BearerAuth, sess.AuthTokenType())
		sess.SetAuthToken(alice.AuthToken())
		assert.Equal(t, "a3859489780243a78b331bd44f58255b552dee104041a45c0e79b610f63af2e5", sess.ID)
		assert.Equal(t, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0", sess.AuthToken())
		assert.Equal(t, header.BearerAuth, sess.AuthTokenType())
	})
}

func TestSession_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := FindSessionByRefID("sessxkkcxxxx")
		assert.Empty(t, m)
		s := &Session{
			UserName:    "charles",
			SessExpires: UnixDay * 3,
			SessTimeout: UnixTime() + UnixWeek,
			RefID:       "sessxkkcxxxx",
		}

		s.SetAuthToken("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7xxx")

		err := s.Create()

		if err != nil {
			t.Fatal(err)
		}

		m2 := FindSessionByRefID("sessxkkcxxxx")
		assert.Equal(t, "charles", m2.UserName)
	})
	t.Run("Invalid RefID", func(t *testing.T) {
		authToken := "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7111"
		id := rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7111")

		m, _ := FindSession(id)

		assert.Empty(t, m)

		s := &Session{
			UserName:    "charles",
			SessExpires: UnixDay * 3,
			SessTimeout: UnixTime() + UnixWeek,
			RefID:       "123",
		}

		s.SetAuthToken(authToken)

		err := s.Create()

		if err != nil {
			t.Fatal(err)
		}

		m2, _ := FindSession(id)

		assert.NotEqual(t, "123", m2.RefID)
	})
	t.Run("ID already exists", func(t *testing.T) {
		authToken := "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0"

		s := &Session{
			UserName:    "charles",
			SessExpires: UnixDay * 3,
			SessTimeout: UnixTime() + UnixWeek,
			RefID:       "sessxkkcxxxx",
		}

		s.SetAuthToken(authToken)

		err := s.Create()
		assert.Error(t, err)
	})
}

func TestSession_Save(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := FindSessionByRefID("sessxkkcxxxy")
		assert.Empty(t, m)
		s := &Session{
			UserName:    "chris",
			SessExpires: UnixDay * 3,
			SessTimeout: UnixTime() + UnixWeek,
			RefID:       "sessxkkcxxxy",
		}

		s.SetAuthToken("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7xxy")

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

func TestSession_SetPreviewToken(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := &Session{ID: "12345678"}
		m.SetPreviewToken("12345")
		assert.Equal(t, "12345", m.PreviewToken)
	})
	t.Run("ID empty", func(t *testing.T) {
		m := &Session{ID: ""}
		m.SetPreviewToken("12345")
		assert.Equal(t, "", m.PreviewToken)
	})
}

func TestSession_SetDownloadToken(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := &Session{ID: "12345678"}
		m.SetDownloadToken("12345")
		assert.Equal(t, "12345", m.DownloadToken)
	})
	t.Run("ID empty", func(t *testing.T) {
		m := &Session{ID: ""}
		m.SetDownloadToken("12345")
		assert.Equal(t, "", m.DownloadToken)
	})
}

func TestSession_IsSuperAdmin(t *testing.T) {
	alice := FindSessionByRefID("sessxkkcabcd")
	alice.RefreshUser()
	assert.True(t, alice.IsSuperAdmin())

	bob := FindSessionByRefID("sessxkkcabce")
	bob.RefreshUser()
	assert.False(t, bob.IsSuperAdmin())

	m := &Session{}
	assert.False(t, m.IsSuperAdmin())

}

func TestSession_NotRegistered(t *testing.T) {
	alice := FindSessionByRefID("sessxkkcabcd")
	alice.RefreshUser()
	assert.False(t, alice.NotRegistered())

	m := &Session{}
	assert.True(t, m.NotRegistered())

}

func TestSession_NoShares(t *testing.T) {
	alice := FindSessionByRefID("sessxkkcabcd")
	alice.RefreshUser()
	alice.User().RefreshShares()
	assert.False(t, alice.NoShares())

	bob := FindSessionByRefID("sessxkkcabce")
	bob.RefreshUser()
	assert.True(t, bob.NoShares())

	m := &Session{}
	assert.True(t, m.NoShares())
}

func TestSession_HasShare(t *testing.T) {
	alice := FindSessionByRefID("sessxkkcabcd")
	alice.RefreshUser()
	alice.User().RefreshShares()
	assert.True(t, alice.HasShare("as6sg6bxpogaaba9"))
	assert.False(t, alice.HasShare("as6sg6bxpogaaba7"))

	bob := FindSessionByRefID("sessxkkcabce")
	bob.RefreshUser()
	bob.User().RefreshShares()
	assert.False(t, bob.HasShare("as6sg6bxpogaaba9"))

	m := &Session{}
	assert.False(t, m.HasShare("as6sg6bxpogaaba9"))
}

func TestSession_SharedUIDs(t *testing.T) {
	alice := FindSessionByRefID("sessxkkcabcd")
	alice.RefreshUser()
	alice.User().RefreshShares()
	assert.Equal(t, "as6sg6bxpogaaba9", alice.SharedUIDs()[0])

	bob := FindSessionByRefID("sessxkkcabce")
	bob.RefreshUser()
	bob.User().RefreshShares()
	assert.Empty(t, bob.SharedUIDs())

	m := &Session{}
	assert.Empty(t, m.SharedUIDs())
}

func TestSession_RedeemToken(t *testing.T) {
	t.Run("bob", func(t *testing.T) {
		bob := FindSessionByRefID("sessxkkcabce")
		bob.RefreshUser()
		bob.User().RefreshShares()
		assert.Equal(t, 0, bob.RedeemToken("1234"))
		assert.Empty(t, bob.User().UserShares)
		assert.Equal(t, 1, bob.RedeemToken("1jxf3jfn2k"))
		bob.User().RefreshShares()
		assert.Equal(t, "as6sg6bxpogaaba8", bob.User().UserShares[0].ShareUID)
	})
	t.Run("Empty session", func(t *testing.T) {
		m := &Session{}
		assert.Equal(t, 0, m.RedeemToken("1234"))
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

func TestSession_SetUserAgent(t *testing.T) {
	t.Run("user agent empty", func(t *testing.T) {
		m := &Session{}
		assert.Equal(t, "", m.UserAgent)
		m.SetUserAgent("")
		assert.Equal(t, "", m.UserAgent)
		m.SetUserAgent("       ")
		assert.Equal(t, "", m.UserAgent)
	})
	t.Run("change user agent", func(t *testing.T) {
		m := &Session{}
		assert.Equal(t, "", m.UserAgent)
		m.SetUserAgent("chrome")
		assert.Equal(t, "chrome", m.UserAgent)
		m.SetUserAgent("mozilla")
		assert.Equal(t, "mozilla", m.UserAgent)
	})
}

func TestSession_SetClientIP(t *testing.T) {
	t.Run("ip empty", func(t *testing.T) {
		m := &Session{}
		assert.Equal(t, "", m.ClientIP)
		m.SetClientIP("")
		assert.Equal(t, "", m.ClientIP)
		m.SetClientIP("       ")
		assert.Equal(t, "", m.ClientIP)
	})
	t.Run("change ip", func(t *testing.T) {
		m := &Session{}
		assert.Equal(t, "", m.ClientIP)
		m.SetClientIP("1234")
		assert.Equal(t, "", m.ClientIP)
		m.SetClientIP("111.123.1.11")
		assert.Equal(t, "111.123.1.11", m.ClientIP)
		m.SetClientIP("2001:db8::68")
		assert.Equal(t, "2001:db8::68", m.ClientIP)
	})
}

func TestSession_HttpStatus(t *testing.T) {
	m := &Session{}
	assert.Equal(t, 401, m.HttpStatus())
	m.Status = 403
	assert.Equal(t, 403, m.HttpStatus())
	alice := FindSessionByRefID("sessxkkcabcd")
	assert.Equal(t, 200, alice.HttpStatus())
}
