package entity

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/time/unix"
)

func TestAuthSession(t *testing.T) {
	t.Run("RandomAppPassword", func(t *testing.T) {
		// Create test request form.
		f := form.Login{
			Username: "alice",
			Password: rnd.AppPassword(),
		}

		// Create test request context.
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(f))
		c.Request.RemoteAddr = "1.2.3.4"

		// Check authentication result.
		authSess, authUser, authErr := AuthSession(f, c)

		assert.Nil(t, authSess)
		assert.Nil(t, authUser)
		assert.Error(t, authErr)
	})
	t.Run("RandomAuthToken", func(t *testing.T) {
		// Create test request form.
		f := form.Login{
			Username: "alice",
			Password: rnd.AuthToken(),
		}

		// Create test request context.
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(f))
		c.Request.RemoteAddr = "1.2.3.4"

		// Check authentication result.
		authSess, authUser, authErr := AuthSession(f, c)

		assert.Nil(t, authSess)
		assert.Nil(t, authUser)
		assert.Error(t, authErr)
	})
	t.Run("AliceAuthToken", func(t *testing.T) {
		s := SessionFixtures.Get("alice_token")

		// Create test request form.
		f := form.Login{
			Username: "alice",
			Password: s.AuthToken(),
		}

		// Create test request context.
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(f))
		c.Request.RemoteAddr = "1.2.3.4"

		// Check authentication result.
		authSess, authUser, authErr := AuthSession(f, c)

		assert.Nil(t, authSess)
		assert.Nil(t, authUser)
		assert.Error(t, authErr)
	})
	t.Run("RemovedRow", func(t *testing.T) {
		token := rnd.AppPassword()
		s := NewSession(3600, 0).SetUser(UserFixtures.Pointer("alice")).SetAuthToken(token)
		s.SetClientIP("10.1.1.1")
		require.NoError(t, s.Save())

		_, err := FindSession(s.ID)
		require.NoError(t, err)
		removeSessionRow(t, s.ID)

		f := form.Login{Username: "alice", Password: token}
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(f))
		c.Request.RemoteAddr = "10.2.2.2:1234"

		authSess, authUser, authErr := AuthSession(f, c)

		assert.Nil(t, authSess)
		assert.Nil(t, authUser)
		assert.ErrorIs(t, authErr, authn.ErrInvalidPassword)
		assert.Equal(t, 0, countSessions(t, s.ID), "the session row must stay deleted")
	})
	t.Run("RemovedRowSameContext", func(t *testing.T) {
		token := rnd.AppPassword()
		s := NewSession(3600, 0).SetUser(UserFixtures.Pointer("alice")).SetAuthToken(token)
		s.SetClientIP("10.3.3.3")
		require.NoError(t, s.Save())

		_, err := FindSession(s.ID)
		require.NoError(t, err)
		removeSessionRow(t, s.ID)

		f := form.Login{Username: "alice", Password: token}
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(f))
		c.Request.RemoteAddr = "10.3.3.3:1234"

		authSess, _, authErr := AuthSession(f, c)

		assert.Nil(t, authSess)
		assert.ErrorIs(t, authErr, authn.ErrInvalidPassword)
	})
	t.Run("AliceTokenPersonal", func(t *testing.T) {
		s := SessionFixtures.Get("alice_token_personal")
		u := FindUserByName("alice")

		// Create test request form.
		f := form.Login{
			Username: "alice",
			Password: s.AuthToken(),
		}

		// Create test request context.
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(f))
		c.Request.RemoteAddr = "1.2.3.4"

		// Check authentication result.
		authSess, authUser, authErr := AuthSession(f, c)

		if authErr != nil {
			t.Fatal(authErr)
		}

		assert.NotNil(t, authSess)
		assert.NotNil(t, authUser)

		assert.Equal(t, u.UserUID, s.UserUID)
		assert.Equal(t, u.Username(), s.GetUserName())
		assert.Equal(t, authUser.UserUID, authSess.UserUID)
		assert.Equal(t, authUser.Username(), authSess.GetUserName())
		assert.Equal(t, authUser.UserUID, authUser.UserUID)
		assert.Equal(t, authUser.Username(), authUser.Username())

		assert.True(t, authSess.IsRegistered())
		assert.True(t, authSess.HasUser())

		assert.True(t, authSess.ValidateScope(acl.ResourceWebDAV, acl.Permissions{acl.ActionCreate}))
		assert.True(t, authSess.ValidateScope(acl.ResourceSessions, acl.Permissions{acl.ActionCreate}))
	})
	t.Run("AliceTokenWebdav", func(t *testing.T) {
		s := SessionFixtures.Get("alice_token_webdav")
		u := FindUserByName("alice")

		// Create test request form.
		f := form.Login{
			Username: "alice",
			Password: s.AuthToken(),
		}

		// Create test request context.
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(f))
		c.Request.RemoteAddr = "1.2.3.4"

		// Check authentication result.
		authSess, authUser, authErr := AuthSession(f, c)

		if authErr != nil {
			t.Fatal(authErr)
		}

		assert.NotNil(t, authSess)
		assert.NotNil(t, authUser)

		assert.Equal(t, u.UserUID, s.UserUID)
		assert.Equal(t, u.Username(), s.GetUserName())
		assert.Equal(t, authUser.UserUID, authSess.UserUID)
		assert.Equal(t, authUser.Username(), authSess.GetUserName())
		assert.Equal(t, authUser.UserUID, authUser.UserUID)
		assert.Equal(t, authUser.Username(), authUser.Username())

		assert.True(t, authSess.IsRegistered())
		assert.True(t, authSess.HasUser())

		assert.True(t, authSess.ValidateScope(acl.ResourceWebDAV, acl.Permissions{acl.ActionCreate}))
		assert.False(t, authSess.ValidateScope(acl.ResourceSessions, acl.Permissions{acl.ActionCreate}))
	})
	t.Run("EmptyPassword", func(t *testing.T) {
		// Create test request form.
		f := form.Login{
			Username: "alice",
			Password: "",
		}

		// Create test request context.
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(f))
		c.Request.RemoteAddr = "1.2.3.4"

		// Check authentication result.
		authSess, authUser, authErr := AuthSession(f, c)

		assert.Nil(t, authSess)
		assert.Nil(t, authUser)
		assert.Error(t, authErr)
	})
}

func TestAuthLocal(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		m := FindSessionByRefID("sessxkkcabch")
		u := FindUserByName("alice")

		// Create test request form.
		frm := form.Login{
			Username: "alice",
			Password: "Alice123!",
		}

		// Create test request context.
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		c.Request.RemoteAddr = "1.2.3.4"

		// Check authentication result.
		if provider, method, err := AuthLocal(u, frm, m, c); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, provider, authn.ProviderLocal)
			assert.Equal(t, method, authn.MethodDefault)
		}
	})
	t.Run("WrongCredentials", func(t *testing.T) {
		m := FindSessionByRefID("sessxkkcabch")
		u := FindUserByName("alice")

		// Create test request form.
		frm := form.Login{
			Username: "alice",
			Password: "photoprism",
		}

		// Create test request context.
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		c.Request.RemoteAddr = "1.2.3.4"

		// Check authentication result.
		if provider, method, err := AuthLocal(u, frm, m, c); err == nil {
			t.Fatal("auth should fail")
		} else {
			assert.Equal(t, provider, authn.ProviderNone)
			assert.Equal(t, method, authn.MethodUndefined)
		}
	})
	t.Run("NoLoginRights", func(t *testing.T) {
		m := &Session{}
		u := FindUserByName("friend")

		u.CanLogin = false

		// Create test request form.
		frm := form.Login{
			Username: "friend",
			Password: "!Friend321",
		}

		// Create test request context.
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		c.Request.RemoteAddr = "1.2.3.4"

		// Check authentication result.
		if provider, method, err := AuthLocal(u, frm, m, c); err == nil {
			t.Fatal("auth should fail")
		} else {
			assert.Equal(t, provider, authn.ProviderNone)
			assert.Equal(t, method, authn.MethodUndefined)
		}

		u.CanLogin = true
	})
	t.Run("AuthenticationDisabled", func(t *testing.T) {
		m := &Session{}
		u := FindUserByName("friend")

		u.SetProvider(authn.ProviderNone)

		// Create test request form.
		frm := form.Login{
			Username: "friend",
			Password: "!Friend321",
		}

		// Create test request context.
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		c.Request.RemoteAddr = "1.2.3.4"

		// Check authentication result.
		if provider, method, err := AuthLocal(u, frm, m, c); err == nil {
			t.Fatal("auth should fail")
		} else {
			assert.Equal(t, provider, authn.ProviderNone)
			assert.Equal(t, method, authn.MethodUndefined)
		}

		u.SetProvider(authn.ProviderLocal)
	})
	t.Run("AliceToken", func(t *testing.T) {
		m := FindSessionByRefID("sess6ey1ykya")
		u := FindUserByName("alice")

		// Create test request form.
		//nolint:gosec // G101: Static password fixture is intentional for login tests.
		frm := form.Login{
			Username: "alice",
			Password: "DIbS8T-uyGMe1-R3fmTv-vVaR35",
		}

		// Create test request context.
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		c.Request.RemoteAddr = "1.2.3.4"

		// Check authentication result.
		if provider, method, err := AuthLocal(u, frm, m, c); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, provider, authn.ProviderApplication)
			assert.Equal(t, method, authn.MethodSession)
		}
	})
	t.Run("AliceTokenInsufficientScope", func(t *testing.T) {
		m := FindSessionByRefID("sesshjtgx8qt")
		u := FindUserByName("alice")

		// Create test request form.
		//nolint:gosec // G101: Static password fixture is intentional for login tests.
		frm := form.Login{
			Username: "alice",
			Password: "5d0rGx-EvsDnV-DcKtYY-HT1aWL",
		}

		// Create test request context.
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		c.Request.RemoteAddr = "1.2.3.4"

		// Check authentication result.
		if provider, method, err := AuthLocal(u, frm, m, c); err == nil {
			t.Fatal("auth should fail")
		} else {
			assert.Equal(t, provider, authn.ProviderNone)
			assert.Equal(t, method, authn.MethodUndefined)
		}
	})
	t.Run("AliceTokenWrongUser", func(t *testing.T) {
		m := FindSessionByRefID("sess6ey1ykya")
		u := FindUserByName("bob")

		// Create test request form.
		//nolint:gosec // G101: Static password fixture is intentional for login tests.
		frm := form.Login{
			Username: "alice",
			Password: "DIbS8T-uyGMe1-R3fmTv-vVaR35",
		}

		// Create test request context.
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		c.Request.RemoteAddr = "1.2.3.4"

		// Check authentication result.
		if provider, method, err := AuthLocal(u, frm, m, c); err == nil {
			t.Fatal("auth should fail")
		} else {
			assert.Equal(t, provider, authn.ProviderNone)
			assert.Equal(t, method, authn.MethodUndefined)
		}
	})
	t.Run("None", func(t *testing.T) {
		m := FindSessionByRefID("sessxkkcaert")
		u := FindUserByName("no_local_auth")

		// Create test request form.
		frm := form.Login{
			Username: "no_local_auth",
			Password: "None123!",
		}

		// Create test request context.
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		c.Request.RemoteAddr = "1.2.3.4"

		// Check authentication result.
		provider, _, err := AuthLocal(u, frm, m, c)
		assert.Error(t, err)
		assert.Equal(t, provider, authn.ProviderNone)
	})
}

func TestSessionLogIn(t *testing.T) {
	const clientIp = "1.2.3.4"
	rec := httptest.NewRecorder()

	t.Run("Admin", func(t *testing.T) {
		m := NewSession(unix.Day, unix.Hour*6)
		m.SetClientIP(clientIp)

		// Create login form.
		frm := form.Login{
			Username: "admin",
			Password: "photoprism",
		}

		// Create test request context.
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		c.Request.RemoteAddr = "1.2.3.4"

		// Test credentials.
		if err := m.LogIn(frm, c); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("Jane", func(t *testing.T) {
		m := NewSession(unix.Day, unix.Hour*6)
		m.SetClientIP(clientIp)

		passcode, codeErr := PasscodeFixtureJane.GenerateCode()

		assert.NoError(t, codeErr)
		assert.Len(t, passcode, 6)

		// Create login form.
		frm := form.Login{
			Username: "jane",
			Password: "Jane123!",
			Code:     passcode,
		}

		// Create test request context.
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		c.Request.RemoteAddr = "1.2.3.4"

		// Test credentials.
		if err := m.LogIn(frm, c); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("InvalidPasscode", func(t *testing.T) {
		m := NewSession(unix.Day, unix.Hour*6)
		m.SetClientIP(clientIp)

		// Create login form.
		frm := form.Login{
			Username: "jane",
			Password: "Jane123!",
			Code:     "xxxxxx",
		}

		// Create test request context.
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		c.Request.RemoteAddr = "1.2.3.4"

		// Expect "passcode required" error after trying to log in.
		err := m.LogIn(frm, c)

		assert.ErrorIs(t, err, authn.ErrInvalidPasscode)
	})
	t.Run("PasscodeRequired", func(t *testing.T) {
		m := NewSession(unix.Day, unix.Hour*6)
		m.SetClientIP(clientIp)

		// Create login form.
		frm := form.Login{
			Username: "jane",
			Password: "Jane123!",
		}

		// Create test request context.
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		c.Request.RemoteAddr = "1.2.3.4"

		// Expect "passcode required" error after trying to log in.
		err := m.LogIn(frm, c)

		assert.ErrorIs(t, err, authn.ErrPasscodeRequired)
	})
	t.Run("InvalidPassword", func(t *testing.T) {
		m := NewSession(unix.Day, unix.Hour*6)
		m.SetClientIP(clientIp)

		// Create login form.
		frm := form.Login{
			Username: "admin",
			Password: "wrong",
		}

		// Create test request context.
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		c.Request.RemoteAddr = "1.2.3.4"

		// Try to log in.
		if err := m.LogIn(frm, c); err == nil {
			t.Fatal("login should fail")
		}
	})
	t.Run("InvalidUser", func(t *testing.T) {
		m := NewSession(unix.Day, unix.Hour*6)
		m.SetClientIP(clientIp)

		// Create login form.
		frm := form.Login{
			Username: "foo",
			Password: "password",
		}

		// Create test request context.
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		c.Request.RemoteAddr = "1.2.3.4"

		// Try to log in.
		if err := m.LogIn(frm, c); err == nil {
			t.Fatal("login should fail")
		}
	})
	t.Run("UnknownUserWithToken", func(t *testing.T) {
		m := NewSession(unix.Day, unix.Hour*6)
		m.SetClientIP(clientIp)

		// Create login form.
		//nolint:gosec // G101: Static token fixture is intentional for login tests.
		frm := form.Login{
			Token: "1jxf3jfn2k",
		}

		// Create test request context.
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		c.Request.RemoteAddr = "1.2.3.4"

		// Try to log in.
		if err := m.LogIn(frm, c); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("TokenRedeemedOnceCountsOneView", func(t *testing.T) {
		link := NewLink("as6sg6bxpogaaba8", false, false)

		if err := link.Save(); err != nil {
			t.Fatal(err)
		}

		before := FindLink(link.LinkUID).LinkViews

		m := NewSession(unix.Day, unix.Hour*6)
		m.SetClientIP(clientIp)

		frm := form.Login{Token: link.LinkToken}

		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		c.Request.RemoteAddr = "1.2.3.4"

		if err := m.LogIn(frm, c); err != nil {
			t.Fatal(err)
		}

		// One login redeems the token once, however often the handler consults it.
		assert.Equal(t, before+1, FindLink(link.LinkUID).LinkViews)
	})
	t.Run("UnknownUserWithInvalidToken", func(t *testing.T) {
		m := NewSession(unix.Day, unix.Hour*6)
		m.SetClientIP(clientIp)

		// Create login form.
		frm := form.Login{
			Token: "1jxf3jfxxx",
		}

		// Create test request context.
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		c.Request.RemoteAddr = "1.2.3.4"

		// Try to log in.
		if err := m.LogIn(frm, c); err == nil {
			t.Fatal("login should fail")
		}
	})
	t.Run("UnknownUserWithoutToken", func(t *testing.T) {
		m := NewSession(unix.Day, unix.Hour*6)
		m.SetClientIP(clientIp)

		// Create login form.
		frm := form.Login{}

		// Create test request context.
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		c.Request.RemoteAddr = "1.2.3.4"

		// Try to log in.
		if err := m.LogIn(frm, c); err == nil {
			t.Fatal("login should fail")
		}
	})
	t.Run("KnownUserWithToken", func(t *testing.T) {
		m := FindSessionByRefID("sessxkkcabch")
		m.SetClientIP(clientIp)

		// Create login form.
		//nolint:gosec // G101: Static token fixture is intentional for login tests.
		frm := form.Login{
			Token: "1jxf3jfn2k",
		}

		// Create test request context.
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		c.Request.RemoteAddr = "1.2.3.4"

		// Try to log in.
		if err := m.LogIn(frm, c); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("KnownUserWithInvalidToken", func(t *testing.T) {
		m := FindSessionByRefID("sessxkkcabch")
		m.SetClientIP(clientIp)

		// Create login form.
		frm := form.Login{
			Token: "1jxf3jfxxx",
		}

		// Create test request context.
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		c.Request.RemoteAddr = "1.2.3.4"

		// Try to log in.
		if err := m.LogIn(frm, c); err == nil {
			t.Fatal("login should fail")
		}
	})
	t.Run("Jane", func(t *testing.T) {
		m := NewSession(unix.Day, unix.Hour*6)
		m.SetClientIP(clientIp)

		passcode, codeErr := PasscodeFixtureJane.GenerateCode()

		assert.NoError(t, codeErr)
		assert.Len(t, passcode, 6)

		// Create login form.
		frm := form.Login{
			Username: "jane",
			Password: "Jane123!",
			Code:     passcode,
		}

		// Create test request context.
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		c.Request.RemoteAddr = "1.2.3.4"

		// Test credentials.
		if err := m.LogIn(frm, c); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSessionLogIn_PreviewToken(t *testing.T) {
	const clientIp = "1.2.3.4"
	const password = "LogIn123!"
	const invalidShareToken = "1jxf3jfxxx"

	// Each account has its own preview token, so no fixture session keeps it admitted.
	newUser := func(t *testing.T, name string) *User {
		u := NewUser()
		u.UserName = name
		u.UserRole = acl.RoleAdmin.String()
		u.AuthProvider = authn.ProviderLocal.String()
		u.CanLogin = true
		require.NoError(t, u.Create())
		require.NoError(t, u.SetPassword(password))
		t.Cleanup(func() {
			_ = u.DeletePassword()
			deleteTestUser(t, u)
		})
		require.NotEmpty(t, u.PreviewToken)
		return u
	}
	newStoredSession := func(t *testing.T, u *User) *Session {
		m := NewSession(unix.Day, unix.Hour*6)
		m.SetClientIP(clientIp)
		m.SetUser(u)
		require.NoError(t, m.Save())
		t.Cleanup(func() { _ = m.Delete() })
		return m
	}
	logIn := func(m *Session, frm form.Login) error {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		c.Request.RemoteAddr = clientIp
		return m.LogIn(frm, c)
	}
	// A login limit without refill shows each charge as a missing token.
	newLoginLimit := func(t *testing.T) {
		prev := limiter.Login
		limiter.Login = limiter.NewLimit(rate.Every(24*time.Hour), 10)
		t.Cleanup(func() { limiter.Login = prev })
	}
	loginTokens := func(m *Session) float64 {
		return limiter.Login.IP(m.IP()).Tokens()
	}

	t.Run("ContinuedInvalidPassword", func(t *testing.T) {
		u := newUser(t, "login-token-a")
		m := newStoredSession(t, u)

		assert.Error(t, logIn(m, form.Login{Username: u.UserName, Password: "wrong-password"}))
		assert.False(t, PreviewToken.Has(m.ID))
		assert.True(t, InvalidPreviewToken(u.PreviewToken))
	})
	t.Run("ContinuedPasscodeRequired", func(t *testing.T) {
		u := FindUserByName("jane")
		require.NotNil(t, u)
		m := newStoredSession(t, u)

		assert.ErrorIs(t, logIn(m, form.Login{Username: "jane", Password: "Jane123!"}), authn.ErrPasscodeRequired)
		assert.False(t, PreviewToken.Has(m.ID))
	})
	t.Run("ContinuedInvalidShareToken", func(t *testing.T) {
		u := newUser(t, "login-token-b")
		m := newStoredSession(t, u)
		id := m.ID
		newLoginLimit(t)

		require.NoError(t, logIn(m, form.Login{Username: u.UserName, Password: password, Token: invalidShareToken}))
		assert.InDelta(t, 9, loginTokens(m), 0.01)
		assert.NotEqual(t, id, m.ID)
		assert.Equal(t, u.UserUID, m.UserUID)
		assert.Equal(t, u.PreviewToken, PreviewToken.Get(m.ID))
		assert.False(t, PreviewToken.Has(id))
	})
	t.Run("NewInvalidShareToken", func(t *testing.T) {
		u := newUser(t, "login-token-c")
		m := NewSession(unix.Day, unix.Hour*6)
		m.SetClientIP(clientIp)
		newLoginLimit(t)

		require.NoError(t, logIn(m, form.Login{Username: u.UserName, Password: password, Token: invalidShareToken}))
		assert.InDelta(t, 9, loginTokens(m), 0.01)
		assert.Equal(t, u.UserUID, m.UserUID)
		assert.Equal(t, u.PreviewToken, PreviewToken.Get(m.ID))
	})
	t.Run("ValidShareToken", func(t *testing.T) {
		u := newUser(t, "login-token-k")
		link := newTestLink(t, 0)
		m := newStoredSession(t, u)
		newLoginLimit(t)

		require.NoError(t, logIn(m, form.Login{Username: u.UserName, Password: password, Token: link.LinkToken}))
		assert.InDelta(t, 10, loginTokens(m), 0.01)
		assert.True(t, m.HasShare(link.ShareUID))
	})
	t.Run("VisitorSignsIn", func(t *testing.T) {
		u := newUser(t, "login-token-l")
		link := newTestLink(t, 0)
		m := NewSession(unix.Day, unix.Hour*6)
		m.SetClientIP(clientIp)

		require.NoError(t, logIn(m, form.Login{Token: link.LinkToken}))
		require.True(t, m.IsVisitor())
		require.NoError(t, m.Save())
		t.Cleanup(func() { _ = m.Delete() })
		visitorToken := m.PreviewToken
		require.NotEmpty(t, visitorToken)
		require.False(t, InvalidPreviewToken(visitorToken))

		require.NoError(t, logIn(m, form.Login{Username: u.UserName, Password: password}))
		assert.Equal(t, u.UserUID, m.UserUID)
		assert.Equal(t, u.PreviewToken, PreviewToken.Get(m.ID))
		assert.True(t, InvalidPreviewToken(visitorToken))
	})
	t.Run("ContinuedOtherAccount", func(t *testing.T) {
		u := newUser(t, "login-token-g")
		other := newUser(t, "login-token-h")
		m := newStoredSession(t, u)

		require.NoError(t, logIn(m, form.Login{Username: other.UserName, Password: password}))
		assert.Equal(t, other.PreviewToken, PreviewToken.Get(m.ID))
		assert.True(t, InvalidPreviewToken(u.PreviewToken))
		assert.False(t, InvalidPreviewToken(other.PreviewToken))
	})
	t.Run("ContinuedOtherAccountInvalidShareToken", func(t *testing.T) {
		u := newUser(t, "login-token-i")
		other := newUser(t, "login-token-j")
		m := newStoredSession(t, u)

		newLoginLimit(t)

		require.NoError(t, logIn(m, form.Login{Username: other.UserName, Password: password, Token: invalidShareToken}))
		assert.InDelta(t, 9, loginTokens(m), 0.01)
		assert.Equal(t, other.PreviewToken, PreviewToken.Get(m.ID))
		assert.True(t, InvalidPreviewToken(u.PreviewToken))
	})
	t.Run("OtherSessionActive", func(t *testing.T) {
		u := newUser(t, "login-token-d")
		newStoredSession(t, u)
		m := newStoredSession(t, u)

		assert.Error(t, logIn(m, form.Login{Username: u.UserName, Password: "wrong-password"}))
		assert.False(t, PreviewToken.Has(m.ID))
		assert.False(t, InvalidPreviewToken(u.PreviewToken))
	})
	t.Run("StoredInvalidShareToken", func(t *testing.T) {
		u := newUser(t, "login-token-e")
		m := newStoredSession(t, u)
		id := m.ID
		newLoginLimit(t)

		assert.Error(t, logIn(m, form.Login{Token: invalidShareToken}))
		assert.InDelta(t, 10, loginTokens(m), 0.01)
		assert.Equal(t, id, m.ID)
		assert.True(t, PreviewToken.Has(m.ID))
		assert.False(t, InvalidPreviewToken(u.PreviewToken))
	})
	t.Run("Success", func(t *testing.T) {
		u := newUser(t, "login-token-f")
		m := newStoredSession(t, u)
		id := m.ID
		newLoginLimit(t)

		require.NoError(t, logIn(m, form.Login{Username: u.UserName, Password: password}))
		assert.InDelta(t, 10, loginTokens(m), 0.01)
		assert.NotEqual(t, id, m.ID)
		assert.True(t, PreviewToken.Has(m.ID))
		assert.False(t, InvalidPreviewToken(u.PreviewToken))
	})
}
