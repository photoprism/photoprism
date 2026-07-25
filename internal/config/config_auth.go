package config

import (
	"crypto/rand"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/photoprism/photoprism/internal/auth/tokens"
	"github.com/photoprism/photoprism/internal/config/ttl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

const (
	// AuthModePublic disables authentication and runs the app in public mode.
	AuthModePublic = "public"
	// AuthModePasswd enables password-based authentication (default).
	AuthModePasswd = "password"
)

func isBcrypt(s string) bool {
	b, err := regexp.MatchString(`^\$2[ayb]\$.{56}$`, s)
	if err != nil {
		return false
	}
	return b
}

// SetAuthMode changes the authentication mode (for use in tests only).
func (c *Config) SetAuthMode(mode string) {
	if !c.Debug() {
		return
	}

	switch mode {
	case AuthModePublic:
		c.options.AuthMode = AuthModePublic
		c.options.Public = true
		entity.ValidateTokens = false
	default:
		c.options.AuthMode = AuthModePasswd
		c.options.Public = false
		entity.ValidateTokens = true
	}
}

// Auth checks if authentication is required.
func (c *Config) Auth() bool {
	return !c.Public()
}

// AuthMode returns the authentication mode.
func (c *Config) AuthMode() string {
	if c.options.Public || c.Demo() {
		return AuthModePublic
	}

	switch c.options.AuthMode {
	case AuthModePublic:
		return AuthModePublic
	default:
		return AuthModePasswd
	}
}

// AuthSecret returns the key for signing authentication tokens, if specified.
func (c *Config) AuthSecret() string {
	return c.options.AuthSecret
}

// Public checks if app runs in public mode and requires no authentication.
func (c *Config) Public() bool {
	return c.AuthMode() == AuthModePublic
}

// AdminUser returns the admin auth name.
func (c *Config) AdminUser() string {
	c.options.AdminUser = clean.Username(c.options.AdminUser)

	if c.options.AdminUser == "" {
		c.options.AdminUser = "admin"
	}

	return c.options.AdminUser
}

// AdminPassword returns the initial admin password.
func (c *Config) AdminPassword() string {
	// Try to read password from file if c.options.AdminPassword is not set.
	if c.options.AdminPassword != "" {
		return clean.Password(c.options.AdminPassword)
	} else if fileName := FlagFilePath("ADMIN_PASSWORD"); fileName == "" {
		// No password set, this is not an error.
		return ""
	} else if b, err := os.ReadFile(fileName); err != nil || len(b) == 0 { //nolint:gosec // path is derived from config directory
		event.SystemWarn([]string{"config", "admin password", "read %s", "%s"}, clean.Log(fileName), clean.Error(err))
		return ""
	} else {
		return clean.Password(string(b))
	}
}

// AdminScope returns the initial admin account scope.
func (c *Config) AdminScope() string {
	if c.options.AdminScope == "" {
		return ""
	}

	return clean.Scope(c.options.AdminScope)
}

// PasswordLength returns the minimum password length in characters.
func (c *Config) PasswordLength() int {
	switch {
	case c.Public():
		return 0
	case c.options.PasswordLength < 1:
		return entity.PasswordLengthDefault
	case c.options.PasswordLength > txt.ClipPassword:
		return txt.ClipPassword
	}

	return c.options.PasswordLength
}

// CheckPassword compares given password p with the admin password
func (c *Config) CheckPassword(p string) bool {
	ap := c.AdminPassword()

	if isBcrypt(ap) {
		err := bcrypt.CompareHashAndPassword([]byte(ap), []byte(p))
		return err == nil
	}

	return ap == p
}

// PasswordResetUri returns the password reset page URI, if any.
func (c *Config) PasswordResetUri() string {
	if c.Public() {
		return ""
	}

	return c.options.PasswordResetUri
}

// RegisterUri returns the user registration page URI, if any.
func (c *Config) RegisterUri() string {
	if c.Public() {
		return ""
	}

	return c.options.RegisterUri
}

// LoginUri returns the user authentication page URI.
func (c *Config) LoginUri() string {
	if c.Public() {
		return c.FrontendUri("/")
	}

	if c.options.LoginUri == "" {
		return c.FrontendUri("/login")
	}

	return c.options.LoginUri
}

// LoginInfo returns the login info text for the page footer.
func (c *Config) LoginInfo() string {
	return c.options.LoginInfo
}

// SessionMaxAge returns the standard session expiration time in seconds.
func (c *Config) SessionMaxAge() int64 {
	if c.options.SessionMaxAge < 0 {
		return 0
	} else if c.options.SessionMaxAge == 0 {
		return DefaultSessionMaxAge
	}

	return c.options.SessionMaxAge
}

// SessionTimeout returns the standard session idle time in seconds.
func (c *Config) SessionTimeout() int64 {
	if c.options.SessionTimeout < 0 {
		return 0
	} else if c.options.SessionTimeout == 0 {
		return DefaultSessionTimeout
	}

	return c.options.SessionTimeout
}

// SessionCache returns the default session cache duration in seconds.
func (c *Config) SessionCache() int64 {
	switch {
	case c.options.SessionCache == 0:
		return DefaultSessionCache
	case c.options.SessionCache < 60:
		return 60
	case c.options.SessionCache > 3600:
		return 3600
	}

	return c.options.SessionCache
}

// SessionCacheDuration returns the default session cache duration.
func (c *Config) SessionCacheDuration() time.Duration {
	return time.Duration(c.SessionCache()) * time.Second
}

// DownloadToken returns the single coarse download token — the admin-configured static value
// (PHOTOPRISM_DOWNLOAD_TOKEN) for permanent, cacheable URLs, or one auto-generated random value when
// none is set. It is delivered to the sessionless public/share client configs and propagated to
// tokens.CoarseDownload; authenticated sessions instead receive a signed, session-scoped token. The
// auto-generated fallback is cached separately so it never overwrites the options, keeping an
// admin-configured static token distinguishable.
func (c *Config) DownloadToken() string {
	if c.Public() {
		return entity.TokenPublic
	} else if c.options.DownloadToken != "" {
		return c.options.DownloadToken
	}

	c.downloadTokenOnce.Do(func() {
		c.downloadToken = rnd.Base36(8)
	})

	return c.downloadToken
}

// TokenSigningKey returns the instance secret that signs the app's URL tokens, generating one on first
// use and nil if that fails, which leaves the signers unconfigured so they refuse.
// One key covers every token kind and never reaches clients. It is kept in memory even when it cannot be
// persisted to config/keys/signing.key, and regenerated when missing, so it is not backed up.
func (c *Config) TokenSigningKey() []byte {
	c.tokenKeyOnce.Do(func() {
		keyPath := filepath.Join(c.KeysPath(), signingKeyName)

		// Reuse the persisted key so tokens stay valid across restarts and replicas.
		if data, err := os.ReadFile(keyPath); err == nil && len(data) >= tokens.KeyLen { //nolint:gosec // path is computed from the config directory
			c.tokenKey = data
			return
		}

		// Leave the key unset on failure: a partially filled buffer would sign with a guessable — worst
		// case all-zero — key, which is strictly worse than rejecting every token.
		key := make([]byte, tokens.KeyLen)
		if _, err := rand.Read(key); err != nil {
			event.SystemError([]string{"config", "token signing key", "generate", "%s"}, clean.Error(err))
			return
		}

		// Keep the key in memory even if it cannot be persisted below.
		c.tokenKey = key

		// Best-effort persistence: a write failure must not clear the in-memory key.
		if err := fs.MkdirAll(c.KeysPath()); err != nil {
			event.SystemWarn([]string{"config", "token signing key", "create keys directory", "%s"}, clean.Error(err))
		} else if err := os.WriteFile(keyPath, key, fs.ModeSecretFile); err != nil {
			event.SystemWarn([]string{"config", "token signing key", "store", "%s"}, clean.Error(err))
		}
	})

	return c.tokenKey
}

// DownloadTokenMaxAge returns the lifetime of signed download tokens, ttl.DownloadTokenDefaultAge
// unless PHOTOPRISM_DOWNLOAD_TOKEN_MAXAGE (seconds) is set.
// The value MUST exceed the client's token-refresh interval so a held token stays valid, so a smaller
// one is raised to ttl.DownloadTokenMinAge and downloads cannot silently break.
func (c *Config) DownloadTokenMaxAge() time.Duration {
	// Read the built-in default, not ttl.DownloadToken, which Propagate overwrites with the result of
	// this call: deriving from it would make the effective lifetime unable to return to the default.
	maxAge := ttl.DownloadTokenDefaultAge.Int()

	if c.options.DownloadTokenMaxAge > 0 {
		maxAge = int(c.options.DownloadTokenMaxAge)
	}

	if maxAge < ttl.DownloadTokenMinAge.Int() {
		maxAge = ttl.DownloadTokenMinAge.Int()
	}

	return time.Duration(maxAge) * time.Second
}

// PreviewToken returns the thumbnail and video streaming API token, derived from the instance signing
// key unless a static one is configured.
// It is not derived from the storage serial, which is world-readable by design so that it survives a
// UID/GID change, and is replaced by signed preview tokens once those land.
func (c *Config) PreviewToken() string {
	if c.Public() {
		return entity.TokenPublic
	} else if c.options.PreviewToken == "" {
		derived := tokens.Derive(c.TokenSigningKey(), tokens.PurposePreview)

		if derived == "" {
			return PreviewTokenPlaceholder
		}

		c.options.PreviewToken = derived
	}

	return c.options.PreviewToken
}

// InvalidPreviewToken checks if the preview token is invalid.
func (c *Config) InvalidPreviewToken(t string) bool {
	return entity.InvalidPreviewToken(t)
}
