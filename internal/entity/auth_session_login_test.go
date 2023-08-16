package entity

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/photoprism/photoprism/pkg/authn"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/form"
)

func TestAuthLocal(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {

		m := FindSessionByRefID("sessxkkcabch")

		u := FindUserByName("alice")

		frm := form.Login{
			UserName: "alice",
			Password: "Alice123!",
		}

		if err := AuthLocal(u, frm, m); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("Wrong credentials", func(t *testing.T) {

		m := FindSessionByRefID("sessxkkcabch")

		u := FindUserByName("alice")

		frm := form.Login{
			UserName: "alice",
			Password: "photoprism",
		}

		if err := AuthLocal(u, frm, m); err == nil {
			t.Fatal("auth should fail")
		}
	})
	t.Run("No login rights", func(t *testing.T) {

		m := &Session{}

		u := FindUserByName("friend")

		u.CanLogin = false

		frm := form.Login{
			UserName: "friend",
			Password: "!Friend321",
		}

		if err := AuthLocal(u, frm, m); err == nil {
			t.Fatal("auth should fail")
		}

		u.CanLogin = true
	})
	t.Run("Authentication disabled", func(t *testing.T) {

		m := &Session{}

		u := FindUserByName("friend")

		u.SetProvider(authn.ProviderNone)

		frm := form.Login{
			UserName: "friend",
			Password: "!Friend321",
		}

		if err := AuthLocal(u, frm, m); err == nil {
			t.Fatal("auth should fail")
		}

		u.SetProvider(authn.ProviderLocal)
	})
}

func TestSessionLogIn(t *testing.T) {
	const clientIp = "1.2.3.4"

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)

	t.Run("Admin", func(t *testing.T) {
		m := NewSession(UnixDay, UnixHour*6)
		m.SetClientIP(clientIp)

		// Create login form.
		frm := form.Login{
			UserName: "admin",
			Password: "photoprism",
		}

		// Create HTTP request.
		ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		ctx.Request.RemoteAddr = "1.2.3.4"

		// Try to log in.
		if err := m.LogIn(frm, ctx); err != nil {
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

		// Create HTTP request.
		ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		ctx.Request.RemoteAddr = "1.2.3.4"

		// Try to log in.
		if err := m.LogIn(frm, ctx); err == nil {
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

		// Create HTTP request.
		ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		ctx.Request.RemoteAddr = "1.2.3.4"

		// Try to log in.
		if err := m.LogIn(frm, ctx); err == nil {
			t.Fatal("login should fail")
		}
	})
	t.Run("Unknown user with token", func(t *testing.T) {
		m := NewSession(UnixDay, UnixHour*6)
		m.SetClientIP(clientIp)

		// Create login form.
		frm := form.Login{
			AuthToken: "1jxf3jfn2k",
		}

		// Create HTTP request.
		ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		ctx.Request.RemoteAddr = "1.2.3.4"

		// Try to log in.
		if err := m.LogIn(frm, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Unknown user with invalid token", func(t *testing.T) {
		m := NewSession(UnixDay, UnixHour*6)
		m.SetClientIP(clientIp)

		// Create login form.
		frm := form.Login{
			AuthToken: "1jxf3jfxxx",
		}

		// Create HTTP request.
		ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		ctx.Request.RemoteAddr = "1.2.3.4"

		// Try to log in.
		if err := m.LogIn(frm, ctx); err == nil {
			t.Fatal("login should fail")
		}
	})

	t.Run("Known user with token", func(t *testing.T) {
		m := FindSessionByRefID("sessxkkcabch")
		m.SetClientIP(clientIp)

		// Create login form.
		frm := form.Login{
			AuthToken: "1jxf3jfn2k",
		}

		// Create HTTP request.
		ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		ctx.Request.RemoteAddr = "1.2.3.4"

		// Try to log in.
		if err := m.LogIn(frm, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Known user with invalid token", func(t *testing.T) {
		m := FindSessionByRefID("sessxkkcabch")
		m.SetClientIP(clientIp)

		// Create login form.
		frm := form.Login{
			AuthToken: "1jxf3jfxxx",
		}

		// Create HTTP request.
		ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", form.AsReader(frm))
		ctx.Request.RemoteAddr = "1.2.3.4"

		// Try to log in.
		if err := m.LogIn(frm, ctx); err == nil {
			t.Fatal("login should fail")
		}
	})
}
