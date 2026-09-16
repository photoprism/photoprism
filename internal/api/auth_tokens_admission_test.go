package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/authn"
)

func TestAuthDownloadAdmission(t *testing.T) {
	t.Run("DisabledAppPasswordsRefuseAnIssuedToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		options := *conf.Options()
		t.Cleanup(func() { *conf.Options() = options; conf.Propagate() })
		conf.SetAuthMode(config.AuthModePasswd)
		t.Cleanup(func() { conf.SetAuthMode(config.AuthModePublic) })
		conf.Options().OriginalsPath = t.TempDir()
		conf.Propagate()
		GetClientConfig(router)
		GetPhotoDownload(router)

		sess, err := entity.AddClientSession("download-admission", 3600, "config photos", authn.GrantPassword, entity.UserFixtures.Pointer("alice"))
		require.NoError(t, err)
		t.Cleanup(func() { _ = sess.Delete() })
		require.True(t, sess.IsApplication())

		const uid = "ps6sg6be2lvl0y13"
		file, err := query.FileByPhotoUID(uid)
		require.NoError(t, err)
		original := CreateTestOriginal(t, file)

		// The client config delivers the token, and the token serves the original bytes.
		cfg := AuthenticatedRequest(app, http.MethodGet, "/api/v1/config", sess.AuthToken())
		require.Equal(t, http.StatusOK, cfg.Code)
		token := gjson.GetBytes(cfg.Body.Bytes(), "downloadToken").String()
		require.NotEmpty(t, token)
		granted := PerformRequest(app, http.MethodGet, "/api/v1/photos/"+uid+"/dl?t="+token)
		require.Equal(t, http.StatusOK, granted.Code)
		require.Equal(t, original, granted.Body.Bytes())

		appPasswords := conf.Settings().Features.AppPasswords
		t.Cleanup(func() { conf.Settings().Features.AppPasswords = appPasswords })
		conf.Settings().Features.AppPasswords = false
		require.True(t, conf.DisableAppPasswords())

		// Turning the feature off refuses the header request and the token alike.
		bearer := AuthenticatedRequest(app, http.MethodGet, "/api/v1/photos/"+uid+"/dl", sess.AuthToken())
		assert.Equal(t, http.StatusForbidden, bearer.Code)
		refused := PerformRequest(app, http.MethodGet, "/api/v1/photos/"+uid+"/dl?t="+token)
		assert.Equal(t, http.StatusForbidden, refused.Code)
		assert.NotEqual(t, original, refused.Body.Bytes())
	})
}

func TestDownloadNotAdmitted(t *testing.T) {
	conf := get.Config()
	conf.SetAuthMode(config.AuthModePasswd)
	t.Cleanup(func() { conf.SetAuthMode(config.AuthModePublic) })

	t.Run("PublicModeAdmitsAnyone", func(t *testing.T) {
		conf.SetAuthMode(config.AuthModePublic)
		defer conf.SetAuthMode(config.AuthModePasswd)
		sess := entity.Session{}
		assert.False(t, downloadNotAdmitted(downloadCtx(""), sess.SetUser(&entity.UnknownUser)))
	})
	t.Run("UserSessionAdmitted", func(t *testing.T) {
		sess := entity.SessionFixtures.Get("alice")
		assert.False(t, downloadNotAdmitted(downloadCtx(""), &sess))
	})
	t.Run("VisitorAdmitted", func(t *testing.T) {
		sess := entity.SessionFixtures.Get("visitor")
		assert.False(t, downloadNotAdmitted(downloadCtx(""), sess.SetUser(&entity.Visitor)))
	})
	t.Run("SessionWithoutUserAdmitted", func(t *testing.T) {
		sess := entity.Session{AuthProvider: authn.ProviderClient.String()}
		assert.False(t, downloadNotAdmitted(downloadCtx(""), &sess))
	})
	t.Run("AppPasswordWithDeniedOwnerRefused", func(t *testing.T) {
		// The feature stays enabled, and the owner is registered and active, so the account's Web
		// UI/API access is the only thing that can refuse this one.
		require.False(t, conf.DisableAppPasswords())
		owner := entity.UserFixtures.Get("alice")
		owner.CanLogin = false
		owner.SuperAdmin = false
		require.False(t, owner.IsUnknown() || owner.IsDisabled() || !owner.IsRegistered())
		sess := entity.Session{AuthProvider: authn.ProviderApplication.String()}
		assert.True(t, downloadNotAdmitted(downloadCtx(""), sess.SetUser(&owner)))
	})
	t.Run("AppPasswordWithoutOwnerRefused", func(t *testing.T) {
		// Pins the order: an app password is checked before the no-account shortcut is taken.
		sess := entity.Session{AuthProvider: authn.ProviderApplication.String()}
		require.True(t, sess.NoUser())
		assert.True(t, downloadNotAdmitted(downloadCtx(""), &sess))
	})
	t.Run("UnknownOwnerRefused", func(t *testing.T) {
		sess := entity.Session{}
		assert.True(t, downloadNotAdmitted(downloadCtx(""), sess.SetUser(&entity.UnknownUser)))
	})
	t.Run("DisabledOwnerRefused", func(t *testing.T) {
		sess := entity.Session{}
		owner := entity.UserFixtures.Get("deleted")
		assert.True(t, downloadNotAdmitted(downloadCtx(""), sess.SetUser(&owner)))
	})
	t.Run("ClientWithUnregisteredOwnerRefused", func(t *testing.T) {
		sess := entity.Session{AuthProvider: authn.ProviderClient.String()}
		assert.True(t, downloadNotAdmitted(downloadCtx(""), sess.SetUser(&entity.Visitor)))
	})
	t.Run("HeaderAuthorizedSessionExempt", func(t *testing.T) {
		// The cluster JWT path is admitted by authAnyJWT, which resolves no account to check.
		sess := entity.Session{}
		sess.SetUser(&entity.UnknownUser)
		c := downloadCtx("")
		assert.True(t, downloadNotAdmitted(c, &sess))
		c.Set(downloadHeaderAuthKey, true)
		assert.False(t, downloadNotAdmitted(c, &sess))
	})
}
