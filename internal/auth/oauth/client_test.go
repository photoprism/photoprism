package oauth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
)

// testAppConfig returns the test config with the app flow enabled.
func testAppConfig(t *testing.T) *config.Config {
	t.Helper()

	conf := config.TestConfig()
	opt := conf.Options()
	apps, public, authMode := opt.OAuthApps, opt.Public, opt.AuthMode
	t.Cleanup(func() { opt.OAuthApps, opt.Public, opt.AuthMode = apps, public, authMode })
	opt.OAuthApps, opt.Public, opt.AuthMode = config.OAuthAppsRegistered, false, config.AuthModePasswd

	return conf
}

// newTestAppClient registers an app client and deletes it after the test.
func newTestAppClient(t *testing.T, scope string, callbacks ...string) *entity.Client {
	t.Helper()

	m, err := entity.AddClient(form.Client{ClientName: "Test App", AuthMethod: authn.MethodPKCE.String(), AuthScope: scope, RedirectURIs: callbacks})
	require.NoError(t, err)
	t.Cleanup(func() { _ = m.Delete() })

	return m
}

func TestFindClient(t *testing.T) {
	conf := testAppConfig(t)
	app := newTestAppClient(t, "mcp photos", "https://claude.ai/api/mcp/auth_callback")

	t.Run("AppClient", func(t *testing.T) {
		c, err := FindClient(conf, app.ClientUID)
		require.NoError(t, err)
		assert.Equal(t, app.ClientUID, c.ID)
		assert.Equal(t, "Test App", c.Name)
		assert.Equal(t, "mcp photos", c.Scope)
		assert.Equal(t, []string{"https://claude.ai/api/mcp/auth_callback"}, c.RedirectURIs)
	})
	t.Run("ClientCredentials", func(t *testing.T) {
		_, err := FindClient(conf, "cs5gfen1bgxz7s9i")
		assert.ErrorIs(t, err, ErrInvalidClient)
	})
	t.Run("Unknown", func(t *testing.T) {
		_, err := FindClient(conf, "csunknown0000001")
		assert.ErrorIs(t, err, ErrInvalidClient)
	})
	t.Run("MetadataURL", func(t *testing.T) {
		_, err := FindClient(conf, "https://claude.ai/oauth/client.json")
		assert.ErrorIs(t, err, ErrInvalidClient)
	})
	t.Run("Disabled", func(t *testing.T) {
		m := newTestAppClient(t, "mcp", "https://claude.ai/cb")
		m.AuthEnabled = false
		require.NoError(t, m.Save())
		_, err := FindClient(conf, m.ClientUID)
		assert.ErrorIs(t, err, ErrInvalidClient)
	})
	t.Run("Cached", func(t *testing.T) {
		m := newTestAppClient(t, "mcp", "https://claude.ai/cb")
		_, err := FindClient(conf, m.ClientUID)
		require.NoError(t, err)

		// A change that bypasses the entity methods is not seen until the cache is flushed.
		require.NoError(t, entity.UnscopedDb().Model(&entity.Client{}).Where("client_uid = ?", m.ClientUID).UpdateColumn("auth_enabled", false).Error)
		_, err = FindClient(conf, m.ClientUID)
		assert.NoError(t, err)
		entity.FlushClientCache()
		_, err = FindClient(conf, m.ClientUID)
		assert.ErrorIs(t, err, ErrInvalidClient)
	})
	t.Run("DisabledAfterRead", func(t *testing.T) {
		m := newTestAppClient(t, "mcp", "https://claude.ai/cb")
		_, err := FindClient(conf, m.ClientUID)
		require.NoError(t, err)
		m.AuthEnabled = false
		require.NoError(t, m.Save())
		_, err = FindClient(conf, m.ClientUID)
		assert.ErrorIs(t, err, ErrInvalidClient)
	})
	t.Run("FlowDisabled", func(t *testing.T) {
		conf.Options().OAuthApps = config.OAuthAppsNone
		_, err := FindClient(conf, app.ClientUID)
		assert.ErrorIs(t, err, ErrInvalidClient)
		conf.Options().OAuthApps = config.OAuthAppsRegistered
	})
}

func TestRedirectOrigin(t *testing.T) {
	t.Run("Web", func(t *testing.T) {
		assert.Equal(t, "https://claude.ai", RedirectOrigin("https://claude.ai/api/mcp/auth_callback"))
		assert.Equal(t, "http://127.0.0.1:33418", RedirectOrigin("http://127.0.0.1:33418/callback"))
	})
	t.Run("Native", func(t *testing.T) {
		assert.Equal(t, "app.photoprism.mobile:", RedirectOrigin("App.PhotoPrism.Mobile:/oauth"))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.Equal(t, "", RedirectOrigin("/relative"))
		assert.Equal(t, "", RedirectOrigin("%zz"))
	})
}
