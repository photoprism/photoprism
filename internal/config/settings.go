package config

import (
	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config/customize"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// initSettings initializes the customization settings from the "settings.yml" file.
func (c *Config) initSettings() {
	if c.settings != nil {
		return
	}

	// Create settings struct.
	c.settings = customize.NewSettings(c.DefaultTheme(), c.DefaultLocale())

	// Get filenames to load the settings from.
	configPath := c.ConfigPath()
	settingsFile := c.SettingsYaml()
	defaultsFile := c.SettingsYamlDefaults(settingsFile)

	// Make sure that the config path exists.
	if err := fs.MkdirAll(configPath); err != nil {
		log.Errorf("settings: %s", createError(configPath, err))
	}

	// Load values from an existing YAML file or create it otherwise.
	if err := c.settings.Load(defaultsFile); err == nil {
		log.Debugf("settings: loaded from %s", defaultsFile)
	} else if err = c.settings.Save(settingsFile); err != nil {
		log.Errorf("settings: could not create %s (%s)", settingsFile, err)
	} else {
		log.Debugf("settings: saved to %s ", settingsFile)
	}

	i18n.SetDir(c.LocalesPath())

	c.settings.Propagate()
}

// Settings returns the global app settings.
func (c *Config) Settings() *customize.Settings {
	// Load settings from the "settings.yml" config file.
	c.initSettings()

	if c.DisablePlaces() {
		c.settings.Features.Places = false
	}

	if c.DisableSettings() {
		c.settings.Features.Settings = false
	}

	if c.DisableFaces() {
		c.settings.Features.People = false
	}

	if c.ReadOnly() {
		c.settings.Features.Upload = false
		c.settings.Features.Import = false
	}

	return c.settings
}

// SessionSettings returns the app settings for the specified session.
func (c *Config) SessionSettings(sess *entity.Session) *customize.Settings {
	// Return global app settings if authentication is disabled (public mode).
	if c.Public() {
		return c.Settings()
	}

	if sess.NoUser() && sess.IsClient() {
		return c.Settings().ApplyACL(acl.Rules, sess.ClientRole()).ApplyScope(sess.Scope())
	}

	user := sess.User()

	// Return public settings if the session does not have a user.
	if user == nil {
		return c.PublicSettings()
	}

	// Apply role-based permissions and user settings to a copy of the global app settings.
	return user.Settings().ApplyTo(c.Settings().ApplyACL(acl.Rules, user.AclRole())).ApplyScope(sess.Scope())
}

// PublicSettings returns the public app settings.
func (c *Config) PublicSettings() *customize.Settings {
	settings := c.Settings()

	return &customize.Settings{
		UI:       settings.UI,
		Search:   settings.Search,
		Maps:     settings.Maps,
		Features: settings.Features,
		Share:    settings.Share,
	}
}

// ShareSettings returns the app settings for share link visitors.
func (c *Config) ShareSettings() *customize.Settings {
	settings := c.Settings().ApplyACL(acl.Rules, acl.RoleVisitor)

	return &customize.Settings{
		UI:       settings.UI,
		Search:   settings.Search,
		Maps:     settings.Maps,
		Features: settings.Features,
		Share:    settings.Share,
	}
}
