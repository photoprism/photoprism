package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
)

// TestCreateSession_AppPassword checks that signing in with an app password follows the app passwords setting.
func TestCreateSession_AppPassword(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)

	CreateSession(router)

	appSess, err := entity.AddClientSession("session-app-password", conf.SessionMaxAge(), "*", authn.GrantPassword, entity.UserFixtures.Pointer("alice"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = appSess.Delete() })

	signIn := func() int {
		return PerformRequestWithBody(app, http.MethodPost, "/api/v1/session", form.AsJson(form.Login{
			Username: "alice",
			Password: appSess.AuthToken(),
		})).Code
	}

	t.Run("Enabled", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, signIn())
	})
	t.Run("Disabled", func(t *testing.T) {
		conf.Settings().Features.AppPasswords = false
		defer func() { conf.Settings().Features.AppPasswords = true }()

		assert.Equal(t, http.StatusForbidden, signIn())
	})
}
