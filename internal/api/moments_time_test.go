package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

func TestGetMomentsTime(t *testing.T) {
	t.Run("PublicMode", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetMomentsTime(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/moments/time")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.LessOrEqual(t, gjson.Get(r.Body.String(), `#(Year=="2790").Count`).Int(), int64(2))
	})
	t.Run("Unauthenticated", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		t.Cleanup(func() { conf.SetAuthMode(config.AuthModePublic) })
		GetMomentsTime(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/moments/time")
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})
}

// TestGetMomentsTime_Access checks that the counts reach only sessions that may see the library,
// and that a restricted scope must include read access to the calendar.
func TestGetMomentsTime_Access(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	t.Cleanup(func() { conf.SetAuthMode(config.AuthModePublic) })
	GetMomentsTime(router)

	session := func(t *testing.T, scope string, user *entity.User) string {
		return scopedSessionToken(t, conf, scope, user)
	}

	request := func(token string) int {
		return AuthenticatedRequest(app, http.MethodGet, "/api/v1/moments/time", token).Code
	}

	t.Run("Admin", func(t *testing.T) {
		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/moments/time", AuthenticateAdmin(app, router))
		assert.Equal(t, http.StatusOK, r.Code)
		assert.True(t, gjson.Get(r.Body.String(), "#").Exists())
	})
	t.Run("Client", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, request(session(t, "calendar", nil)))
	})
	t.Run("Guest", func(t *testing.T) {
		assert.Equal(t, http.StatusForbidden, request(session(t, "*", entity.UserFixtures.Pointer("guest"))))
	})
	t.Run("Visitor", func(t *testing.T) {
		visitor := entity.SessionFixtures.Get("visitor")
		assert.Equal(t, http.StatusForbidden, request(visitor.AuthToken()))
	})
	t.Run("ScopeRead", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, request(session(t, "calendar read", entity.UserFixtures.Pointer("alice"))))
	})
	t.Run("ScopeWrite", func(t *testing.T) {
		assert.Equal(t, http.StatusForbidden, request(session(t, "calendar write", entity.UserFixtures.Pointer("alice"))))
	})
	t.Run("ScopeOtherResource", func(t *testing.T) {
		assert.Equal(t, http.StatusForbidden, request(session(t, "photos", entity.UserFixtures.Pointer("alice"))))
	})
	t.Run("LibraryWithoutSearch", func(t *testing.T) {
		// No built-in role sees the library without search access, so the test installs one.
		role := acl.RoleGuest
		prev, found := acl.Rules[acl.ResourceCalendar][role]
		acl.Rules[acl.ResourceCalendar][role] = acl.Grant{acl.AccessLibrary: true, acl.ActionView: true}
		t.Cleanup(func() {
			if found {
				acl.Rules[acl.ResourceCalendar][role] = prev
			} else {
				delete(acl.Rules[acl.ResourceCalendar], role)
			}
		})

		assert.Equal(t, http.StatusForbidden, request(session(t, "*", entity.UserFixtures.Pointer("guest"))))
	})
}
