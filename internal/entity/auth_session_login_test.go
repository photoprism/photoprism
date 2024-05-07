package entity

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/unix"
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
		assert.Equal(t, u.Username(), s.Username())
		assert.Equal(t, authUser.UserUID, authSess.UserUID)
		assert.Equal(t, authUser.Username(), authSess.Username())
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
		assert.Equal(t, u.Username(), s.Username())
		assert.Equal(t, authUser.UserUID, authSess.UserUID)
		assert.Equal(t, authUser.Username(), authSess.Username())
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
