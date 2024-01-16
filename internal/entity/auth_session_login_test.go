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
)

func TestAuthSession(t *testing.T) {
	t.Run("RandomAuthSecret", func(t *testing.T) {
		// Create test request form.
		f := form.Login{
			UserName: "alice",
			Password: rnd.AuthSecret(),
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
			UserName: "alice",
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
			UserName: "alice",
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
			UserName: "alice",
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

		assert.True(t, authSess.HasScope(acl.ResourceWebDAV.String()))
		assert.True(t, authSess.HasScope(acl.ResourceSessions.String()))
	})
	t.Run("AliceTokenWebdav", func(t *testing.T) {
		s := SessionFixtures.Get("alice_token_webdav")
		u := FindUserByName("alice")

		// Create test request form.
		f := form.Login{
			UserName: "alice",
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

		assert.True(t, authSess.HasScope(acl.ResourceWebDAV.String()))
		assert.False(t, authSess.HasScope(acl.ResourceSessions.String()))
	})
	t.Run("EmptyPassword", func(t *testing.T) {
		// Create test request form.
		f := form.Login{
			UserName: "alice",
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
			UserName: "alice",
			Password: "Alice123!",
		}

		// Create test request context.
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		c.Request.RemoteAddr = "1.2.3.4"

		// Check authentication result.
		if provider, err := AuthLocal(u, frm, m, c); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, authn.ProviderLocal, provider)
		}
	})
	t.Run("Wrong credentials", func(t *testing.T) {
		m := FindSessionByRefID("sessxkkcabch")
		u := FindUserByName("alice")

		// Create test request form.
		frm := form.Login{
			UserName: "alice",
			Password: "photoprism",
		}

		// Create test request context.
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		c.Request.RemoteAddr = "1.2.3.4"

		// Check authentication result.
		if provider, err := AuthLocal(u, frm, m, c); err == nil {
			t.Fatal("auth should fail")
		} else {
			assert.Equal(t, authn.ProviderNone, provider)
		}
	})
	t.Run("No login rights", func(t *testing.T) {
		m := &Session{}
		u := FindUserByName("friend")

		u.CanLogin = false

		// Create test request form.
		frm := form.Login{
			UserName: "friend",
			Password: "!Friend321",
		}

		// Create test request context.
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		c.Request.RemoteAddr = "1.2.3.4"

		// Check authentication result.
		if provider, err := AuthLocal(u, frm, m, c); err == nil {
			t.Fatal("auth should fail")
		} else {
			assert.Equal(t, authn.ProviderNone, provider)
		}

		u.CanLogin = true
	})
	t.Run("Authentication disabled", func(t *testing.T) {
		m := &Session{}
		u := FindUserByName("friend")

		u.SetProvider(authn.ProviderNone)

		// Create test request form.
		frm := form.Login{
			UserName: "friend",
			Password: "!Friend321",
		}

		// Create test request context.
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		c.Request.RemoteAddr = "1.2.3.4"

		// Check authentication result.
		if provider, err := AuthLocal(u, frm, m, c); err == nil {
			t.Fatal("auth should fail")
		} else {
			assert.Equal(t, authn.ProviderNone, provider)
		}

		u.SetProvider(authn.ProviderLocal)
	})
}

func TestSessionLogIn(t *testing.T) {
	const clientIp = "1.2.3.4"
	rec := httptest.NewRecorder()

	t.Run("Admin", func(t *testing.T) {
		m := NewSession(UnixDay, UnixHour*6)
		m.SetClientIP(clientIp)

		// Create login form.
		frm := form.Login{
			UserName: "admin",
			Password: "photoprism",
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
	t.Run("WrongPassword", func(t *testing.T) {
		m := NewSession(UnixDay, UnixHour*6)
		m.SetClientIP(clientIp)

		// Create login form.
		frm := form.Login{
			UserName: "admin",
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
		m := NewSession(UnixDay, UnixHour*6)
		m.SetClientIP(clientIp)

		// Create login form.
		frm := form.Login{
			UserName: "foo",
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
	t.Run("Unknown user with token", func(t *testing.T) {
		m := NewSession(UnixDay, UnixHour*6)
		m.SetClientIP(clientIp)

		// Create login form.
		frm := form.Login{
			ShareToken: "1jxf3jfn2k",
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

	t.Run("Unknown user with invalid token", func(t *testing.T) {
		m := NewSession(UnixDay, UnixHour*6)
		m.SetClientIP(clientIp)

		// Create login form.
		frm := form.Login{
			ShareToken: "1jxf3jfxxx",
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

	t.Run("Known user with token", func(t *testing.T) {
		m := FindSessionByRefID("sessxkkcabch")
		m.SetClientIP(clientIp)

		// Create login form.
		frm := form.Login{
			ShareToken: "1jxf3jfn2k",
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

	t.Run("Known user with invalid token", func(t *testing.T) {
		m := FindSessionByRefID("sessxkkcabch")
		m.SetClientIP(clientIp)

		// Create login form.
		frm := form.Login{
			ShareToken: "1jxf3jfxxx",
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
}
