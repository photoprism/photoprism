package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/customize"
	"github.com/photoprism/photoprism/internal/entity"
)

func TestConfig_ClientConfig(t *testing.T) {
	t.Run("TestConfig", func(t *testing.T) {
		c := TestConfig()
		result := c.ClientPublic()
		assert.IsType(t, ClientConfig{}, result)
		assert.Equal(t, AuthModePublic, result.AuthMode)
		assert.Equal(t, "/library/browse", result.LoginUri)
		assert.Equal(t, "", result.RegisterUri)
		assert.Equal(t, 0, result.PasswordLength)
		assert.Equal(t, "", result.PasswordResetUri)
		assert.Equal(t, true, result.Public)
	})
	t.Run("TestErrorConfig", func(t *testing.T) {
		c := NewTestErrorConfig()
		result2 := c.ClientPublic()
		assert.IsType(t, ClientConfig{}, result2)
		assert.Equal(t, AuthModePasswd, result2.AuthMode)
		assert.Equal(t, false, result2.Public)
	})
	t.Run("Values", func(t *testing.T) {
		c := TestConfig()

		cfg := c.ClientRole(acl.RoleAdmin)

		assert.IsType(t, ClientConfig{}, cfg)

		if cfg.JsUri == "" {
			t.Error("the JavaScript asset URI must not be empty, make sure that the frontend has been built")
		}

		if cfg.CssUri == "" {
			t.Error("the CSS asset URI must not be empty, make sure that the frontend has been built")
		}

		assert.NotEmpty(t, cfg.Name)
		assert.NotEmpty(t, cfg.Version)
		assert.NotEmpty(t, cfg.Copyright)
		assert.NotEmpty(t, cfg.Thumbs)
		assert.NotEmpty(t, cfg.ManifestUri)
		assert.Equal(t, true, cfg.Debug)
		assert.Equal(t, AuthModePublic, cfg.AuthMode)
		assert.Equal(t, false, cfg.Demo)
		assert.Equal(t, true, cfg.Sponsor)
		assert.Equal(t, false, cfg.ReadOnly)

		// Counts.
		assert.NotEmpty(t, cfg.Count.All)
		assert.NotEmpty(t, cfg.Count.Photos)
		assert.LessOrEqual(t, 20, cfg.Count.Photos)
		assert.LessOrEqual(t, 1, cfg.Count.Live)
		assert.LessOrEqual(t, 4, cfg.Count.Videos)
		assert.LessOrEqual(t, cfg.Count.Photos+cfg.Count.Live+cfg.Count.Videos-cfg.Count.Review, cfg.Count.All)
		assert.LessOrEqual(t, 6, cfg.Count.Cameras)
		assert.LessOrEqual(t, 1, cfg.Count.Lenses)
		assert.LessOrEqual(t, 13, cfg.Count.Review)
		assert.LessOrEqual(t, 1, cfg.Count.Private)
		assert.LessOrEqual(t, 4, cfg.Count.Albums)
	})
}

func TestConfig_ClientShareConfig(t *testing.T) {
	config := TestConfig()
	result := config.ClientShare()
	assert.IsType(t, ClientConfig{}, result)
	assert.Equal(t, true, result.Public)
	assert.Equal(t, AuthModePublic, result.AuthMode)
	assert.Equal(t, true, result.Experimental)
	assert.Equal(t, false, result.ReadOnly)
}

func TestConfig_ClientRoleConfig(t *testing.T) {
	c := NewTestConfig("config")
	c.SetAuthMode(AuthModePasswd)

	assert.Equal(t, AuthModePasswd, c.AuthMode())

	adminFeatures := c.ClientRole(acl.RoleAdmin).Settings.Features

	t.Run("RoleAdmin", func(t *testing.T) {
		cfg := c.ClientRole(acl.RoleAdmin)
		assert.IsType(t, ClientConfig{}, cfg)
		assert.Equal(t, AuthModePasswd, cfg.AuthMode)
		assert.Equal(t, false, cfg.Public)

		f := cfg.Settings.Features
		assert.Equal(t, adminFeatures, f)

		expected := customize.FeatureSettings{
			Account:   true,
			Albums:    true,
			Archive:   true,
			Delete:    true,
			Download:  true,
			Edit:      true,
			Estimates: true,
			Favorites: true,
			Files:     true,
			Folders:   true,
			Import:    true,
			Labels:    true,
			Library:   true,
			Logs:      true,
			Moments:   true,
			People:    true,
			Places:    true,
			Private:   true,
			Ratings:   true,
			Reactions: true,
			Review:    true,
			Search:    true,
			Settings:  true,
			Share:     true,
			Services:  true,
			Upload:    true,
			Videos:    true,
		}

		assert.Equal(t, expected, f)
	})
	t.Run("RoleGuest", func(t *testing.T) {
		cfg := c.ClientRole(acl.RoleGuest)
		f := cfg.Settings.Features

		assert.NotEqual(t, adminFeatures, f)

		expected := customize.FeatureSettings{
			Account:   true,
			Albums:    true,
			Archive:   false,
			Delete:    false,
			Download:  true,
			Edit:      false,
			Estimates: true,
			Favorites: false,
			Files:     false,
			Folders:   true,
			Import:    false,
			Labels:    false,
			Library:   false,
			Logs:      false,
			Moments:   true,
			People:    false,
			Places:    true,
			Private:   false,
			Ratings:   false,
			Reactions: true,
			Review:    true,
			Search:    true,
			Settings:  true,
			Share:     false,
			Services:  false,
			Upload:    false,
			Videos:    true,
		}

		assert.Equal(t, expected, f)
	})
	t.Run("RoleVisitor", func(t *testing.T) {
		cfg := c.ClientRole(acl.RoleVisitor)
		f := cfg.Settings.Features

		assert.NotEqual(t, adminFeatures, f)

		expected := customize.FeatureSettings{
			Account:   false,
			Albums:    true,
			Archive:   false,
			Delete:    false,
			Download:  true,
			Edit:      false,
			Estimates: true,
			Favorites: false,
			Files:     false,
			Folders:   true,
			Import:    false,
			Labels:    false,
			Library:   false,
			Logs:      false,
			Moments:   true,
			People:    false,
			Places:    true,
			Private:   false,
			Ratings:   false,
			Reactions: false,
			Review:    true,
			Search:    false,
			Settings:  false,
			Share:     false,
			Services:  false,
			Upload:    false,
			Videos:    false,
		}

		assert.Equal(t, expected, f)
	})
	t.Run("RoleNone", func(t *testing.T) {
		cfg := c.ClientRole(acl.RoleNone)
		f := cfg.Settings.Features

		assert.NotEqual(t, adminFeatures, f)
		assert.False(t, f.Search)
		assert.False(t, f.Videos)
		assert.False(t, f.Albums)
		assert.False(t, f.Moments)
		assert.False(t, f.Labels)
		assert.False(t, f.People)
		assert.False(t, f.Settings)
		assert.False(t, f.Edit)
		assert.False(t, f.Private)
		assert.False(t, f.Upload)
		assert.False(t, f.Download)
		assert.False(t, f.Services)
		assert.False(t, f.Delete)
		assert.False(t, f.Import)
		assert.False(t, f.Library)
		assert.False(t, f.Logs)
		assert.True(t, f.Review)
		assert.False(t, f.Share)
		assert.False(t, f.Favorites)
		assert.False(t, f.Reactions)
		assert.False(t, f.Ratings)
	})
}

func TestConfig_ClientSessionConfig(t *testing.T) {
	c := NewTestConfig("config")
	c.SetAuthMode(AuthModePasswd)

	assert.Equal(t, AuthModePasswd, c.AuthMode())

	adminFeatures := c.ClientRole(acl.RoleAdmin).Settings.Features

	t.Run("RoleAdmin", func(t *testing.T) {
		cfg := c.ClientSession(entity.SessionFixtures.Pointer("alice"))
		assert.IsType(t, ClientConfig{}, cfg)
		assert.Equal(t, false, cfg.Public)
		assert.NotEmpty(t, cfg.PreviewToken)
		assert.NotEmpty(t, cfg.DownloadToken)

		f := cfg.Settings.Features
		assert.Equal(t, adminFeatures, f)

		assert.True(t, f.Search)
		assert.True(t, f.Videos)
		assert.True(t, f.Albums)
		assert.True(t, f.Moments)
		assert.True(t, f.Labels)
		assert.True(t, f.People)
		assert.True(t, f.Settings)
		assert.True(t, f.Edit)
		assert.True(t, f.Private)
		assert.True(t, f.Upload)
		assert.True(t, f.Download)
		assert.True(t, f.Services)
		assert.True(t, f.Delete)
		assert.True(t, f.Import)
		assert.True(t, f.Library)
		assert.True(t, f.Logs)
		assert.True(t, f.Review)
		assert.True(t, f.Share)
	})
	t.Run("RoleAdminToken", func(t *testing.T) {
		cfg := c.ClientSession(entity.SessionFixtures.Pointer("alice_token"))

		assert.IsType(t, ClientConfig{}, cfg)
		assert.Equal(t, false, cfg.Public)
		assert.NotEmpty(t, cfg.PreviewToken)
		assert.NotEmpty(t, cfg.DownloadToken)

		f := cfg.Settings.Features
		assert.Equal(t, adminFeatures, f)

		assert.True(t, f.Search)
		assert.True(t, f.Videos)
		assert.True(t, f.Albums)
		assert.True(t, f.Moments)
		assert.True(t, f.Labels)
		assert.True(t, f.People)
		assert.True(t, f.Settings)
		assert.True(t, f.Edit)
		assert.True(t, f.Private)
		assert.True(t, f.Upload)
		assert.True(t, f.Download)
		assert.True(t, f.Services)
		assert.True(t, f.Delete)
		assert.True(t, f.Import)
		assert.True(t, f.Library)
		assert.True(t, f.Logs)
		assert.True(t, f.Review)
		assert.True(t, f.Share)
	})
	t.Run("RoleAdminTokenScope", func(t *testing.T) {
		cfg := c.ClientSession(entity.SessionFixtures.Pointer("alice_token_scope"))

		assert.IsType(t, ClientConfig{}, cfg)
		assert.Equal(t, false, cfg.Public)
		assert.NotEmpty(t, cfg.PreviewToken)
		assert.NotEmpty(t, cfg.DownloadToken)

		f := cfg.Settings.Features
		assert.NotEqual(t, adminFeatures, f)

		assert.True(t, f.Search)
		assert.True(t, f.Videos)
		assert.True(t, f.Albums)
		assert.False(t, f.Moments)
		assert.False(t, f.Labels)
		assert.False(t, f.People)
		assert.False(t, f.Settings)
		assert.True(t, f.Edit)
		assert.True(t, f.Private)
		assert.True(t, f.Upload)
		assert.True(t, f.Download)
		assert.False(t, f.Services)
		assert.True(t, f.Delete)
		assert.False(t, f.Import)
		assert.False(t, f.Library)
		assert.False(t, f.Logs)
		assert.True(t, f.Review)
		assert.False(t, f.Share)
	})
	t.Run("RoleVisitor", func(t *testing.T) {
		cfg := c.ClientSession(entity.SessionFixtures.Pointer("visitor"))

		assert.IsType(t, ClientConfig{}, cfg)
		assert.Equal(t, false, cfg.Public)
		assert.NotEmpty(t, cfg.PreviewToken)
		assert.NotEmpty(t, cfg.DownloadToken)

		f := cfg.Settings.Features
		assert.NotEqual(t, adminFeatures, f)

		assert.False(t, f.Search)
		assert.False(t, f.Videos)
		assert.True(t, f.Albums)
		assert.True(t, f.Moments)
		assert.True(t, f.Folders)
		assert.False(t, f.Labels)
		assert.False(t, f.People)
		assert.False(t, f.Settings)
		assert.False(t, f.Edit)
		assert.False(t, f.Private)
		assert.False(t, f.Upload)
		assert.True(t, f.Download)
		assert.False(t, f.Services)
		assert.False(t, f.Delete)
		assert.False(t, f.Import)
		assert.False(t, f.Library)
		assert.False(t, f.Logs)
		assert.True(t, f.Review)
		assert.False(t, f.Share)
	})
	t.Run("RoleVisitorTokenMetrics", func(t *testing.T) {
		cfg := c.ClientSession(entity.SessionFixtures.Pointer("visitor_token_metrics"))

		assert.IsType(t, ClientConfig{}, cfg)
		assert.Equal(t, false, cfg.Public)
		assert.NotEmpty(t, cfg.PreviewToken)
		assert.NotEmpty(t, cfg.DownloadToken)

		f := cfg.Settings.Features
		assert.NotEqual(t, adminFeatures, f)

		assert.False(t, f.Search)
		assert.False(t, f.Videos)
		assert.True(t, f.Albums)
		assert.True(t, f.Moments)
		assert.True(t, f.Folders)
		assert.False(t, f.Labels)
		assert.False(t, f.People)
		assert.False(t, f.Settings)
		assert.False(t, f.Edit)
		assert.False(t, f.Private)
		assert.False(t, f.Upload)
		assert.True(t, f.Download)
		assert.False(t, f.Services)
		assert.False(t, f.Delete)
		assert.False(t, f.Import)
		assert.False(t, f.Library)
		assert.False(t, f.Logs)
		assert.True(t, f.Review)
		assert.False(t, f.Share)
	})
	t.Run("RoleNone", func(t *testing.T) {
		sess := entity.SessionFixtures.Pointer("unauthorized")

		cfg := c.ClientSession(sess)

		assert.IsType(t, ClientConfig{}, cfg)
		assert.Equal(t, false, cfg.Public)
		assert.NotEmpty(t, cfg.PreviewToken)
		assert.NotEmpty(t, cfg.DownloadToken)

		f := cfg.Settings.Features
		assert.NotEqual(t, adminFeatures, f)

		assert.False(t, f.Search)
		assert.False(t, f.Videos)
		assert.False(t, f.Albums)
		assert.False(t, f.Moments)
		assert.False(t, f.Labels)
		assert.False(t, f.People)
		assert.False(t, f.Settings)
		assert.False(t, f.Edit)
		assert.False(t, f.Private)
		assert.False(t, f.Upload)
		assert.False(t, f.Download)
		assert.False(t, f.Services)
		assert.False(t, f.Delete)
		assert.False(t, f.Import)
		assert.False(t, f.Library)
		assert.False(t, f.Logs)
		assert.True(t, f.Review)
		assert.False(t, f.Share)
	})
	t.Run("Bob", func(t *testing.T) {
		cfg := c.ClientSession(entity.SessionFixtures.Pointer("bob"))

		assert.IsType(t, ClientConfig{}, cfg)
		assert.Equal(t, false, cfg.Public)
		assert.NotEmpty(t, cfg.PreviewToken)
		assert.NotEmpty(t, cfg.DownloadToken)
		f := cfg.Settings.Features

		assert.True(t, f.Search)
		assert.True(t, f.Videos)
		assert.True(t, f.Albums)
		assert.True(t, f.Moments)
		assert.True(t, f.Labels)
		assert.True(t, f.People)
		assert.True(t, f.Settings)
		assert.True(t, f.Edit)
		assert.True(t, f.Private)
		assert.True(t, f.Upload)
		assert.True(t, f.Download)
		assert.True(t, f.Services)
		assert.True(t, f.Delete)
		assert.True(t, f.Import)
		assert.True(t, f.Library)
		assert.True(t, f.Logs)
		assert.True(t, f.Review)
		assert.True(t, f.Share)
	})
	t.Run("TokenMetrics", func(t *testing.T) {
		cfg := c.ClientSession(entity.SessionFixtures.Pointer("token_metrics"))

		assert.IsType(t, ClientConfig{}, cfg)
		assert.Equal(t, false, cfg.Public)
		assert.NotEmpty(t, cfg.PreviewToken)
		assert.NotEmpty(t, cfg.DownloadToken)

		f := cfg.Settings.Features
		assert.NotEqual(t, adminFeatures, f)

		assert.False(t, f.Search)
		assert.False(t, f.Videos)
		assert.False(t, f.Albums)
		assert.False(t, f.Moments)
		assert.False(t, f.Folders)
		assert.False(t, f.Labels)
		assert.False(t, f.People)
		assert.False(t, f.Settings)
		assert.False(t, f.Edit)
		assert.False(t, f.Private)
		assert.False(t, f.Upload)
		assert.False(t, f.Download)
		assert.False(t, f.Services)
		assert.False(t, f.Delete)
		assert.False(t, f.Import)
		assert.False(t, f.Library)
		assert.False(t, f.Logs)
		assert.True(t, f.Review)
		assert.False(t, f.Share)
	})
	t.Run("TokenSettings", func(t *testing.T) {
		cfg := c.ClientSession(entity.SessionFixtures.Pointer("token_settings"))

		assert.IsType(t, ClientConfig{}, cfg)
		assert.Equal(t, false, cfg.Public)
		assert.NotEmpty(t, cfg.PreviewToken)
		assert.NotEmpty(t, cfg.DownloadToken)

		f := cfg.Settings.Features
		assert.NotEqual(t, adminFeatures, f)

		assert.True(t, f.Search)
		assert.True(t, f.Videos)
		assert.True(t, f.Albums)
		assert.True(t, f.Moments)
		assert.True(t, f.Labels)
		assert.True(t, f.People)
		assert.True(t, f.Settings)
		assert.True(t, f.Edit)
		assert.True(t, f.Private)
		assert.True(t, f.Upload)
		assert.True(t, f.Download)
		assert.False(t, f.Services)
		assert.True(t, f.Delete)
		assert.True(t, f.Import)
		assert.True(t, f.Library)
		assert.True(t, f.Logs)
		assert.True(t, f.Review)
		assert.False(t, f.Share)
	})
}

func TestConfig_Flags(t *testing.T) {
	config := TestConfig()
	config.options.Experimental = true
	config.options.ReadOnly = true
	config.settings.UI.Scrollbar = false

	result := config.Flags()
	assert.Equal(t, []string{"public", "debug", "test", "sponsor", "experimental", "readonly", "settings", "hide-scrollbar"}, result)

	config.options.Experimental = false
	config.options.ReadOnly = false
}
