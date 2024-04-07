package entity

import (
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/header"
	"github.com/photoprism/photoprism/pkg/report"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/unix"
)

func TestNewSession(t *testing.T) {
	t.Run("NoSessionData", func(t *testing.T) {
		m := NewSession(unix.Day, unix.Hour*6)

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
		m := NewSession(unix.Day, unix.Hour*6)
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
		m := NewSession(unix.Day, unix.Hour*6)
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
		m := NewSession(unix.Day, unix.Hour*6)

		assert.NotNil(t, m)

		sess := m.SetData(nil)

		assert.NotNil(t, sess)
		assert.NotEmpty(t, sess.ID)
		assert.Equal(t, sess.ID, m.ID)
	})
}

func TestSession_Expires(t *testing.T) {
	t.Run("Set expiry date", func(t *testing.T) {
		m := NewSession(unix.Day, unix.Hour)
		initialExpiryDate := m.SessExpires
		m.Expires(time.Date(2035, 01, 15, 12, 30, 0, 0, time.UTC))
		finalExpiryDate := m.SessExpires
		assert.Greater(t, finalExpiryDate, initialExpiryDate)

	})
	t.Run("Try to set zero date", func(t *testing.T) {
		m := NewSession(unix.Day, unix.Hour)
		initialExpiryDate := m.SessExpires
		m.Expires(time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC))
		finalExpiryDate := m.SessExpires
		assert.Equal(t, finalExpiryDate, initialExpiryDate)
	})
}

func TestDeleteExpiredSessions(t *testing.T) {
	assert.Equal(t, 0, DeleteExpiredSessions())
	m := NewSession(unix.Day, unix.Hour)
	m.Expires(time.Date(2000, 01, 15, 12, 30, 0, 0, time.UTC))
	m.Save()
	assert.Equal(t, 1, DeleteExpiredSessions())
}

func TestDeleteClientSessions(t *testing.T) {
	// Test client UID.
	clientUID := "cs5gfen1bgx00000"

	// Create new test client.
	client := NewClient()
	client.ClientUID = clientUID

	// Make sure no sessions exist yet and test missing arguments.
	assert.Equal(t, 0, DeleteClientSessions(&Client{}, authn.MethodUndefined, -1))
	assert.Equal(t, 0, DeleteClientSessions(client, authn.MethodOAuth2, -1))
	assert.Equal(t, 0, DeleteClientSessions(client, authn.MethodOAuth2, 0))
	assert.Equal(t, 0, DeleteClientSessions(&Client{}, authn.MethodDefault, 0))

	// Create 10 test client sessions.
	for i := 0; i < 10; i++ {
		sess := NewSession(3600, 0)
		sess.SetClient(client)

		if err := sess.Save(); err != nil {
			t.Fatal(err)
		}
	}

	// Check if the expected number of sessions is deleted until none are left.
	assert.Equal(t, 0, DeleteClientSessions(client, authn.MethodOAuth2, -1))
	assert.Equal(t, 0, DeleteClientSessions(client, authn.MethodDefault, 1))
	assert.Equal(t, 9, DeleteClientSessions(client, authn.MethodOAuth2, 1))
	assert.Equal(t, 1, DeleteClientSessions(client, authn.MethodOAuth2, 0))
	assert.Equal(t, 0, DeleteClientSessions(client, authn.MethodOAuth2, 0))
	assert.Equal(t, 0, DeleteClientSessions(client, authn.MethodUndefined, 0))
}

func TestSessionStatusUnauthorized(t *testing.T) {
	m := SessionStatusUnauthorized()
	assert.Equal(t, http.StatusUnauthorized, m.Status)
	assert.IsType(t, &Session{}, m)
}

func TestSessionStatusForbidden(t *testing.T) {
	m := SessionStatusForbidden()
	assert.Equal(t, http.StatusForbidden, m.Status)
	assert.IsType(t, &Session{}, m)
}

func TestSessionStatusTooManyRequests(t *testing.T) {
	m := SessionStatusTooManyRequests()
	assert.Equal(t, http.StatusTooManyRequests, m.Status)
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
		m := NewSession(unix.Day, unix.Hour)
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
		assert.Equal(t, header.AuthBearer, sess.AuthTokenType())
		sess.Regenerate()
		assert.True(t, rnd.IsSessionID(sess.ID))
		assert.True(t, rnd.IsAuthToken(sess.AuthToken()))
		assert.Equal(t, header.AuthBearer, sess.AuthTokenType())
		sess.SetAuthToken(alice.AuthToken())
		assert.Equal(t, "a3859489780243a78b331bd44f58255b552dee104041a45c0e79b610f63af2e5", sess.ID)
		assert.Equal(t, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0", sess.AuthToken())
		assert.Equal(t, header.AuthBearer, sess.AuthTokenType())
	})
	t.Run("Alice", func(t *testing.T) {
		sess := SessionFixtures.Get("alice")
		assert.Equal(t, "a3859489780243a78b331bd44f58255b552dee104041a45c0e79b610f63af2e5", sess.ID)
		assert.Equal(t, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0", sess.AuthToken())
		assert.Equal(t, header.AuthBearer, sess.AuthTokenType())
	})
	t.Run("Find", func(t *testing.T) {
		alice := SessionFixtures.Get("alice")
		sess := FindSessionByRefID("sessxkkcabcd")
		assert.Equal(t, "a3859489780243a78b331bd44f58255b552dee104041a45c0e79b610f63af2e5", sess.ID)
		assert.Equal(t, "", sess.AuthToken())
		assert.Equal(t, header.AuthBearer, sess.AuthTokenType())
		sess.SetAuthToken(alice.AuthToken())
		assert.Equal(t, "a3859489780243a78b331bd44f58255b552dee104041a45c0e79b610f63af2e5", sess.ID)
		assert.Equal(t, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0", sess.AuthToken())
		assert.Equal(t, header.AuthBearer, sess.AuthTokenType())
	})
}

func TestSession_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := FindSessionByRefID("sessxkkcxxxx")
		assert.Empty(t, m)
		s := &Session{
			UserName:    "charles",
			SessExpires: unix.Day * 3,
			SessTimeout: unix.Time() + unix.Week,
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
			SessExpires: unix.Day * 3,
			SessTimeout: unix.Time() + unix.Week,
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
			SessExpires: unix.Day * 3,
			SessTimeout: unix.Time() + unix.Week,
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
			SessExpires: unix.Day * 3,
			SessTimeout: unix.Time() + unix.Week,
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

	if err := m.Updates(Session{UserName: "anton"}); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "anton", m.UserName)
}

func TestSession_Client(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		m := FindSessionByRefID("sessxkkcabcd")
		assert.Equal(t, "uqxetse3cy5eo9z2", m.UserUID)
		assert.Equal(t, "uqxetse3cy5eo9z2", m.User().UserUID)
		assert.Equal(t, "", m.Client().ClientUID)
		assert.Equal(t, "uqxetse3cy5eo9z2", m.Client().UserUID)
		assert.Equal(t, acl.RoleNone, m.Client().AclRole())
		assert.Equal(t, acl.RoleNone, m.ClientRole())
	})
	t.Run("AliceTokenPersonal", func(t *testing.T) {
		m := SessionFixtures.Get("alice_token_personal")
		assert.Equal(t, "uqxetse3cy5eo9z2", m.UserUID)
		assert.Equal(t, "uqxetse3cy5eo9z2", m.User().UserUID)
		assert.Equal(t, "", m.Client().ClientUID)
		assert.Equal(t, "uqxetse3cy5eo9z2", m.Client().UserUID)
		assert.Equal(t, acl.RoleClient, m.Client().AclRole())
		assert.Equal(t, acl.RoleClient, m.ClientRole())
	})
	t.Run("ClientMetrics", func(t *testing.T) {
		m := SessionFixtures.Get("client_metrics")
		assert.Equal(t, "", m.UserUID)
		assert.Equal(t, "", m.User().UserUID)
		assert.Equal(t, "cs5cpu17n6gj2qo5", m.Client().ClientUID)
		assert.Equal(t, "", m.Client().UserUID)
		assert.Equal(t, acl.RoleClient, m.Client().AclRole())
		assert.Equal(t, acl.RoleClient, m.ClientRole())
	})
	t.Run("Default", func(t *testing.T) {
		m := &Session{}
		assert.Equal(t, "", m.UserUID)
		assert.Equal(t, "", m.User().UserUID)
		assert.Equal(t, "", m.Client().ClientUID)
		assert.Equal(t, "", m.Client().UserUID)
		assert.Equal(t, acl.RoleNone, m.Client().AclRole())
		assert.Equal(t, acl.RoleNone, m.ClientRole())
	})
}

func TestSession_ClientRole(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		m := SessionFixtures.Get("alice")
		assert.Equal(t, acl.RoleNone, m.ClientRole())
	})
	t.Run("AliceTokenPersonal", func(t *testing.T) {
		m := SessionFixtures.Get("alice_token_personal")
		assert.Equal(t, acl.RoleClient, m.ClientRole())
	})
	t.Run("TokenMetrics", func(t *testing.T) {
		m := SessionFixtures.Get("token_metrics")
		assert.Equal(t, acl.RoleClient, m.ClientRole())
	})
	t.Run("TokenSettings", func(t *testing.T) {
		m := SessionFixtures.Get("token_settings")
		assert.Equal(t, acl.RoleClient, m.ClientRole())
	})
	t.Run("Default", func(t *testing.T) {
		m := &Session{}
		assert.Equal(t, acl.RoleNone, m.ClientRole())
	})
}

func TestSession_ClientInfo(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		m := SessionFixtures.Get("alice")
		assert.Equal(t, "n/a", m.ClientInfo())
	})
	t.Run("Metrics", func(t *testing.T) {
		m := SessionFixtures.Get("client_metrics")
		assert.Equal(t, "cs5cpu17n6gj2qo5", m.ClientInfo())
	})
}

func TestSession_NoClient(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		m := SessionFixtures.Get("alice")
		assert.True(t, m.NoClient())
	})
	t.Run("Metrics", func(t *testing.T) {
		m := SessionFixtures.Get("client_metrics")
		assert.False(t, m.NoClient())
	})
}

func TestSession_SetClient(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		m := SessionFixtures.Get("alice")
		assert.Equal(t, acl.RoleNone, m.ClientRole())
		assert.Equal(t, "", m.Client().ClientUID)
		m.SetClient(ClientFixtures.Pointer("alice"))
		assert.Equal(t, acl.RoleClient, m.ClientRole())
		assert.Equal(t, "cs5gfen1bgxz7s9i", m.Client().ClientUID)
	})
}

func TestSession_SetClientName(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		m := SessionFixtures.Get("alice_token_personal")
		assert.Equal(t, "", m.ClientUID)
		assert.Equal(t, "alice_token_personal", m.ClientName)
		assert.Equal(t, "alice_token_personal", m.ClientInfo())
		m.SetClientName("Foo Bar!")
		assert.Equal(t, "", m.ClientUID)
		assert.Equal(t, "Foo Bar!", m.ClientName)
		assert.Equal(t, "Foo Bar!", m.ClientInfo())
		m.SetClientName("")
		assert.Equal(t, "Foo Bar!", m.ClientName)
		assert.Equal(t, "Foo Bar!", m.ClientInfo())
	})
	t.Run("setNewID", func(t *testing.T) {
		m := NewSession(0, 0)
		assert.Equal(t, "", m.ClientUID)
		assert.Equal(t, "", m.ClientName)
		assert.Equal(t, report.NotAssigned, m.ClientInfo())
		m.SetClientName("Foo Bar!")
		assert.Equal(t, "", m.ClientUID)
		assert.Equal(t, "Foo Bar!", m.ClientName)
		assert.Equal(t, "Foo Bar!", m.ClientInfo())
	})
}

func TestSession_User(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		m := FindSessionByRefID("sessxkkcabcd")
		assert.Equal(t, "uqxetse3cy5eo9z2", m.User().UserUID)
	})
	t.Run("Default", func(t *testing.T) {
		m := &Session{}
		assert.Equal(t, "", m.User().UserUID)
	})
}

func TestSession_UserInfo(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		m := SessionFixtures.Get("alice")
		assert.Equal(t, "alice", m.UserInfo())
	})
	t.Run("Metrics", func(t *testing.T) {
		m := SessionFixtures.Get("client_metrics")
		assert.Equal(t, "", m.UserInfo())
	})
}

func TestSession_UserRole(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		m := FindSessionByRefID("sessxkkcabcd")
		assert.Equal(t, acl.RoleAdmin, m.UserRole())
	})
	t.Run("Bob", func(t *testing.T) {
		m := FindSessionByRefID("sessxkkcabce")
		assert.Equal(t, acl.RoleAdmin, m.UserRole())
	})
	t.Run("Default", func(t *testing.T) {
		m := &Session{}
		assert.Equal(t, acl.RoleNone, m.UserRole())
	})
}

func TestSession_RefreshUser(t *testing.T) {
	t.Run("Bob", func(t *testing.T) {
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

func TestSession_AuthInfo(t *testing.T) {
	t.Run("bob", func(t *testing.T) {
		m := FindSessionByRefID("sessxkkcabce")

		i := m.AuthInfo()

		assert.Equal(t, "Default", i)
	})
	t.Run("aliceTokenWebDAV", func(t *testing.T) {
		m := FindSessionByRefID("sesshjtgx8qt")

		i := m.AuthInfo()

		assert.Equal(t, "Access Token", i)
	})
}

func TestSession_SetAuthID(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		s := &Session{
			UserName: "test",
			RefID:    "sessxkkcxxxz",
			AuthID:   "test-session-auth-id",
		}

		m := s.SetAuthID("")

		assert.Equal(t, "test-session-auth-id", m.AuthID)
	})
	t.Run("New", func(t *testing.T) {
		s := &Session{
			UserName: "test",
			RefID:    "sessxkkcxxxz",
			AuthID:   "new-id",
		}

		m := s.SetAuthID("new-id")

		assert.Equal(t, "new-id", m.AuthID)
	})
}

func TestSession_SetMethod(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		s := &Session{
			UserName:     "test",
			RefID:        "sessxkkcxxxz",
			AuthProvider: authn.ProviderAccessToken.String(),
			AuthMethod:   authn.MethodDefault.String(),
		}

		m := s.SetMethod("")

		assert.Equal(t, authn.ProviderAccessToken, m.Provider())
		assert.Equal(t, authn.MethodDefault, m.Method())
	})
	t.Run("Test", func(t *testing.T) {
		s := &Session{
			UserName:     "test",
			RefID:        "sessxkkcxxxz",
			AuthProvider: authn.ProviderAccessToken.String(),
			AuthMethod:   authn.MethodDefault.String(),
		}

		m := s.SetMethod("Test")

		assert.Equal(t, authn.ProviderAccessToken, m.Provider())
		assert.Equal(t, authn.Method("Test"), m.Method())
	})
	t.Run("Test", func(t *testing.T) {
		s := &Session{
			UserName:     "test",
			RefID:        "sessxkkcxxxz",
			AuthProvider: authn.ProviderAccessToken.String(),
			AuthMethod:   authn.MethodDefault.String(),
		}

		m := s.SetMethod(authn.MethodSession)

		assert.Equal(t, authn.ProviderAccessToken, m.Provider())
		assert.Equal(t, authn.MethodSession, m.Method())
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

func TestSession_ValidateScope(t *testing.T) {
	t.Run("AnyScope", func(t *testing.T) {
		s := &Session{
			UserName:  "test",
			RefID:     "sessxkkcxxxz",
			AuthScope: "*",
		}

		assert.True(t, s.ValidateScope("", nil))
	})
	t.Run("ReadScope", func(t *testing.T) {
		s := &Session{
			UserName:  "test",
			RefID:     "sessxkkcxxxz",
			AuthScope: "read",
		}

		assert.True(t, s.ValidateScope("metrics", nil))
		assert.True(t, s.ValidateScope("sessions", nil))
		assert.True(t, s.ValidateScope("metrics", acl.Permissions{acl.ActionView, acl.AccessAll}))
		assert.False(t, s.ValidateScope("metrics", acl.Permissions{acl.ActionUpdate}))
		assert.False(t, s.ValidateScope("metrics", acl.Permissions{acl.ActionUpdate}))
		assert.False(t, s.ValidateScope("settings", acl.Permissions{acl.ActionUpdate}))
		assert.False(t, s.ValidateScope("settings", acl.Permissions{acl.ActionCreate}))
		assert.False(t, s.ValidateScope("sessions", acl.Permissions{acl.ActionDelete}))
	})
	t.Run("ReadAny", func(t *testing.T) {
		s := &Session{
			UserName:  "test",
			RefID:     "sessxkkcxxxz",
			AuthScope: "read *",
		}

		assert.True(t, s.ValidateScope("metrics", nil))
		assert.True(t, s.ValidateScope("sessions", nil))
		assert.True(t, s.ValidateScope("metrics", acl.Permissions{acl.ActionView, acl.AccessAll}))
		assert.False(t, s.ValidateScope("metrics", acl.Permissions{acl.ActionUpdate}))
		assert.False(t, s.ValidateScope("metrics", acl.Permissions{acl.ActionUpdate}))
		assert.False(t, s.ValidateScope("settings", acl.Permissions{acl.ActionUpdate}))
		assert.False(t, s.ValidateScope("settings", acl.Permissions{acl.ActionCreate}))
		assert.False(t, s.ValidateScope("sessions", acl.Permissions{acl.ActionDelete}))
	})
	t.Run("ReadSettings", func(t *testing.T) {
		s := &Session{
			UserName:  "test",
			RefID:     "sessxkkcxxxz",
			AuthScope: "read settings",
		}

		assert.True(t, s.ValidateScope("settings", acl.Permissions{acl.ActionView}))
		assert.False(t, s.ValidateScope("metrics", nil))
		assert.False(t, s.ValidateScope("sessions", nil))
		assert.False(t, s.ValidateScope("metrics", acl.Permissions{acl.ActionView, acl.AccessAll}))
		assert.False(t, s.ValidateScope("metrics", acl.Permissions{acl.ActionUpdate}))
		assert.False(t, s.ValidateScope("metrics", acl.Permissions{acl.ActionUpdate}))
		assert.False(t, s.ValidateScope("settings", acl.Permissions{acl.ActionUpdate}))
		assert.False(t, s.ValidateScope("sessions", acl.Permissions{acl.ActionDelete}))
		assert.False(t, s.ValidateScope("sessions", acl.Permissions{acl.ActionDelete}))
	})
}

func TestSession_InsufficientScope(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		s := &Session{
			UserName:  "test",
			RefID:     "sessxkkcxxxz",
			AuthScope: "*",
		}

		assert.False(t, s.InsufficientScope("", nil))
	})
	t.Run("ReadSettings", func(t *testing.T) {
		s := &Session{
			UserName:  "test",
			RefID:     "sessxkkcxxxz",
			AuthScope: "read settings",
		}

		assert.False(t, s.InsufficientScope("settings", acl.Permissions{acl.ActionView}))
		assert.True(t, s.InsufficientScope("metrics", nil))
		assert.True(t, s.InsufficientScope("sessions", nil))
		assert.True(t, s.InsufficientScope("metrics", acl.Permissions{acl.ActionView, acl.AccessAll}))
		assert.True(t, s.InsufficientScope("metrics", acl.Permissions{acl.ActionUpdate}))
		assert.True(t, s.InsufficientScope("metrics", acl.Permissions{acl.ActionUpdate}))
		assert.True(t, s.InsufficientScope("settings", acl.Permissions{acl.ActionUpdate}))
		assert.True(t, s.InsufficientScope("sessions", acl.Permissions{acl.ActionDelete}))
		assert.True(t, s.InsufficientScope("sessions", acl.Permissions{acl.ActionDelete}))
	})
}

func TestSession_SetScope(t *testing.T) {
	t.Run("EmptyScope", func(t *testing.T) {
		s := &Session{
			UserName:  "test",
			RefID:     "sessxkkcxxxz",
			AuthScope: "*",
		}

		m := s.SetScope("")

		assert.Equal(t, "*", m.AuthScope)
	})
	t.Run("NewScope", func(t *testing.T) {
		s := &Session{
			UserName:  "test",
			RefID:     "sessxkkcxxxz",
			AuthScope: "*",
		}

		m := s.SetScope("Metrics")

		assert.Equal(t, "metrics", m.AuthScope)
	})
}

func TestSession_SetGrantType(t *testing.T) {
	t.Run("Password", func(t *testing.T) {
		m := &Session{
			UserName:  "test",
			RefID:     "sessxkkcxxxz",
			AuthScope: "*",
		}

		expected := "password"

		m.SetGrantType(authn.GrantPassword)
		assert.Equal(t, expected, m.GrantType)
		m.SetGrantType(authn.GrantClientCredentials)
		assert.Equal(t, expected, m.GrantType)
		m.SetGrantType(authn.GrantUndefined)
		assert.Equal(t, expected, m.GrantType)
		assert.Equal(t, authn.GrantPassword, m.AuthGrantType())
	})
	t.Run("ClientCredentials", func(t *testing.T) {
		client := ClientFixtures.Pointer("alice")
		m := client.NewSession(&gin.Context{}, authn.GrantClientCredentials)

		expected := "client_credentials"

		assert.Equal(t, expected, m.GrantType)
		m.SetGrantType(authn.GrantPassword)
		assert.Equal(t, expected, m.GrantType)
		m.SetGrantType(authn.GrantUndefined)
		assert.Equal(t, expected, m.GrantType)
		assert.Equal(t, authn.GrantClientCredentials, m.AuthGrantType())
	})
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

func TestSession_NoUser(t *testing.T) {
	alice := FindSessionByRefID("sessxkkcabcd")
	assert.False(t, alice.NoUser())

	visitor := FindSessionByRefID("sessxkkcabcg")
	assert.False(t, visitor.NoUser())

	metrics := FindSessionByRefID("sessgh6gjuo1")
	assert.True(t, metrics.NoUser())
}

func TestSession_HasRegisteredUser(t *testing.T) {
	alice := FindSessionByRefID("sessxkkcabcd")
	assert.True(t, alice.HasRegisteredUser())

	visitor := FindSessionByRefID("sessxkkcabcg")
	assert.False(t, visitor.HasRegisteredUser())

	metrics := FindSessionByRefID("sessgh6gjuo1")
	assert.False(t, metrics.HasRegisteredUser())
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
		m := NewSession(unix.Day, unix.Hour)
		assert.False(t, m.TimeoutAt().IsZero())
		assert.Equal(t, m.ExpiresAt(), m.TimeoutAt())
		assert.False(t, m.TimedOut())
		assert.Greater(t, m.ExpiresIn(), int64(0))
	})
	t.Run("NoExpiration", func(t *testing.T) {
		m := NewSession(0, unix.Hour)
		t.Logf("Timeout: %s, Expiration: %s", m.TimeoutAt().String(), m.ExpiresAt())
		assert.True(t, m.TimeoutAt().IsZero())
		assert.Equal(t, m.ExpiresAt(), m.TimeoutAt())
		assert.False(t, m.TimedOut())
		assert.True(t, m.ExpiresAt().IsZero())
		assert.Equal(t, m.ExpiresIn(), int64(0))
	})
	t.Run("NoTimeout", func(t *testing.T) {
		m := NewSession(unix.Day, 0)
		t.Logf("Timeout: %s, Expiration: %s", m.TimeoutAt().String(), m.ExpiresAt())
		assert.False(t, m.TimeoutAt().IsZero())
		assert.Equal(t, m.ExpiresAt(), m.TimeoutAt())
		assert.False(t, m.TimedOut())
		assert.False(t, m.ExpiresAt().IsZero())
		assert.Greater(t, m.ExpiresIn(), int64(0))
	})
	t.Run("TimedOut", func(t *testing.T) {
		m := NewSession(unix.Day, unix.Hour)
		utc := unix.Time()

		m.LastActive = utc - (unix.Hour + 1)

		assert.False(t, m.TimeoutAt().IsZero())
		assert.True(t, m.TimedOut())
	})
	t.Run("NotTimedOut", func(t *testing.T) {
		m := NewSession(unix.Day, unix.Hour)
		utc := unix.Time()

		m.LastActive = utc - (unix.Hour - 10)

		assert.False(t, m.TimeoutAt().IsZero())
		assert.False(t, m.TimedOut())
	})
}

func TestSession_Expired(t *testing.T) {
	t.Run("NewSession", func(t *testing.T) {
		m := NewSession(unix.Day, unix.Hour)
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
		m := NewSession(unix.Day, unix.Hour)
		t.Logf("Timeout: %s, Expiration: %s", m.TimeoutAt().String(), m.ExpiresAt())
		utc := unix.Time()

		m.SessExpires = utc - 10

		assert.False(t, m.ExpiresAt().IsZero())
		assert.True(t, m.Expired())
		assert.False(t, m.TimeoutAt().IsZero())
		assert.True(t, m.TimedOut())
		assert.Equal(t, m.ExpiresAt(), m.TimeoutAt())
	})
	t.Run("NotExpired", func(t *testing.T) {
		m := NewSession(unix.Day, unix.Hour)
		utc := unix.Time()

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
