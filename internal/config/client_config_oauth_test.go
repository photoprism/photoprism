package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/tokens"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
)

func TestConfig_ClientOAuth(t *testing.T) {
	c := NewTestConfig("config")
	c.SetAuthMode(AuthModePasswd)
	t.Cleanup(func() { c.SetAuthMode(AuthModePublic) })

	prevCdn, prevVideo := c.options.CdnUrl, c.options.CdnVideo
	c.options.CdnUrl = "https://cdn.example.com/"
	c.options.CdnVideo = true
	t.Cleanup(func() { c.options.CdnUrl, c.options.CdnVideo = prevCdn, prevVideo; c.Propagate() })
	c.Propagate()

	require.True(t, c.UseCdn())
	require.NotEqual(t, c.ApiUri(), c.ContentUri())
	require.NotEqual(t, c.ApiUri(), c.VideoUri())

	// clientSession returns an unsaved session of a new client without an account that has the scope.
	clientSession := func(scope string) *entity.Session {
		client, err := entity.AddClient(form.Client{ClientName: "Client OAuth " + scope, AuthScope: scope})
		require.NoError(t, err)
		t.Cleanup(func() { _ = client.Purge() })

		sess := entity.NewSession(3600, 0)
		sess.SetClient(client)
		require.True(t, sess.SignedPreview())

		return sess
	}

	t.Run("ApiHosts", func(t *testing.T) {
		cfg := c.ClientOAuth(clientSession("*"))
		assert.Equal(t, c.ApiUri(), cfg.ContentUri)
		assert.Equal(t, c.ApiUri(), cfg.VideoUri)
		assert.Empty(t, cfg.PreviewToken)
		assert.Empty(t, cfg.DownloadToken)
	})
	t.Run("SignedPreviewToken", func(t *testing.T) {
		sess := clientSession("*")
		cfg := c.ClientSession(sess)
		assert.Equal(t, c.ApiUri(), cfg.ContentUri)
		assert.Equal(t, c.ApiUri(), cfg.VideoUri)

		expires, sid, token, ok := tokens.ParseCompact(cfg.PreviewToken)
		require.True(t, ok, cfg.PreviewToken)
		assert.Equal(t, sess.ID, sid)
		assert.Equal(t, sess.ExpiresAt().Unix(), expires)
		_, valid := tokens.VerifyPreview(expires, sid, token)
		assert.True(t, valid)
		assert.Empty(t, sess.PreviewToken)

		_, sid, _, ok = tokens.ParseCompact(cfg.DownloadToken)
		require.True(t, ok, cfg.DownloadToken)
		assert.Equal(t, sess.ID, sid)
		assert.NotEqual(t, cfg.PreviewToken, cfg.DownloadToken)
	})
	t.Run("StoredRandomToken", func(t *testing.T) {
		// A session issued with a random token receives the signed token as well.
		sess := clientSession("*")
		sess.PreviewToken = "oldrnd01"
		cfg := c.ClientSession(sess)
		_, sid, _, ok := tokens.ParseCompact(cfg.PreviewToken)
		require.True(t, ok, cfg.PreviewToken)
		assert.Equal(t, sess.ID, sid)
	})
	t.Run("ScopeWithoutPreviews", func(t *testing.T) {
		cfg := c.ClientSession(clientSession("metrics"))
		assert.Empty(t, cfg.PreviewToken)
		assert.Empty(t, cfg.DownloadToken)
		assert.Equal(t, c.ApiUri(), cfg.ContentUri)
	})
	t.Run("ScopeWithPhotos", func(t *testing.T) {
		cfg := c.ClientSession(clientSession("photos"))
		assert.NotEmpty(t, cfg.PreviewToken)
		assert.NotEmpty(t, cfg.DownloadToken)
	})
	t.Run("NoExpiry", func(t *testing.T) {
		sess := clientSession("*")
		sess.SessExpires = -1
		cfg := c.ClientSession(sess)
		expires, _, _, ok := tokens.ParseCompact(cfg.PreviewToken)
		require.True(t, ok, cfg.PreviewToken)
		assert.Equal(t, int64(0), expires%tokens.PreviewDay)
	})
	t.Run("RoleWithoutPreviews", func(t *testing.T) {
		client, err := entity.AddClient(form.Client{ClientName: "Client OAuth None", AuthScope: "*", ClientRole: "none"})
		require.NoError(t, err)
		t.Cleanup(func() { _ = client.Purge() })
		sess := entity.NewSession(3600, 0)
		sess.SetClient(client)
		require.True(t, sess.SignedPreview())
		assert.Empty(t, c.ClientSession(sess).PreviewToken)
	})
	t.Run("JWT", func(t *testing.T) {
		sess := &entity.Session{AuthProvider: authn.ProviderClient.String(), AuthMethod: authn.MethodJWT.String(), AuthScope: "*"}
		cfg := c.ClientSession(sess)
		assert.Empty(t, cfg.PreviewToken)
		assert.Empty(t, cfg.DownloadToken)
		assert.Equal(t, c.ApiUri(), cfg.ContentUri)
	})
	t.Run("AccessTokenWithoutClient", func(t *testing.T) {
		cfg := c.ClientSession(entity.SessionFixtures.Pointer("token_metrics"))
		assert.Empty(t, cfg.PreviewToken)
		assert.Empty(t, cfg.DownloadToken)
	})
	t.Run("Account", func(t *testing.T) {
		sess := entity.SessionFixtures.Pointer("alice")
		cfg := c.ClientSession(sess)
		assert.Equal(t, sess.PreviewToken, cfg.PreviewToken)
		assert.Equal(t, c.ContentUri(), cfg.ContentUri)
		assert.Equal(t, c.VideoUri(), cfg.VideoUri)
	})
	t.Run("AppPassword", func(t *testing.T) {
		sess := entity.SessionFixtures.Pointer("alice_app_password_full_access")
		cfg := c.ClientSession(sess)
		assert.Equal(t, sess.PreviewToken, cfg.PreviewToken)
		assert.Equal(t, c.ContentUri(), cfg.ContentUri)
	})
	t.Run("PublicMode", func(t *testing.T) {
		c.SetAuthMode(AuthModePublic)
		t.Cleanup(func() { c.SetAuthMode(AuthModePasswd) })
		cfg := c.ClientSession(clientSession("*"))
		assert.Equal(t, entity.TokenPublic, cfg.PreviewToken)
		assert.Equal(t, entity.TokenPublic, cfg.DownloadToken)
	})
}
