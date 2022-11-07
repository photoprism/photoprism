package entity

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/form"
)

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
}
