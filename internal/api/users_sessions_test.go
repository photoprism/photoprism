package api

import (
	"net/http"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestFindUserSessions(t *testing.T) {
	t.Run("Public", func(t *testing.T) {
		app, router, _ := NewApiTest()
		FindUserSessions(router)
		r := PerformRequest(app, "GET", "/api/v1/users/uqxc08w3d0ej2283/sessions")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("Unauthorized", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		FindUserSessions(router)

		r := PerformRequest(app, "GET", "/api/v1/users/uqxc08w3d0ej2283/sessions")
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})
	t.Run("InvalidUID", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		FindUserSessions(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		r := AuthenticatedRequest(app, "GET", "/api/v1/users/uqxetseacy5eo9z2/sessions?count=100", sessId)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("Success", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		FindUserSessions(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		r := AuthenticatedRequest(app, "GET", "/api/v1/users/uqxetse3cy5eo9z2/sessions?count=100", sessId)
		assert.Equal(t, http.StatusOK, r.Code)
		// t.Logf("FindUserSessions/Success: %s", r.Body.String())
	})
}
