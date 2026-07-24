package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/tokens"
	"github.com/photoprism/photoprism/internal/config/ttl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

func TestAuth(t *testing.T) {
	c := NewConfig(CliTestContext())
	c.options.Public = true
	c.options.Demo = false
	assert.False(t, c.Auth())
	c.options.Public = false
	c.options.Demo = false
	assert.True(t, c.Auth())
	c.options.Demo = true
	assert.False(t, c.Auth())
}

func TestAuthMode(t *testing.T) {
	c := NewConfig(CliTestContext())
	c.options.Public = true
	c.options.Demo = false
	assert.Equal(t, AuthModePublic, c.AuthMode())
	c.options.Public = false
	c.options.Demo = false
	assert.Equal(t, AuthModePasswd, c.AuthMode())
	c.options.Demo = true
	assert.Equal(t, AuthModePublic, c.AuthMode())
	c.options.AuthMode = "pass"
	assert.Equal(t, AuthModePublic, c.AuthMode())
	c.options.Demo = false
	c.options.AuthMode = "pass"
	assert.Equal(t, AuthModePasswd, c.AuthMode())
	c.options.AuthMode = "password"
	assert.Equal(t, AuthModePasswd, c.AuthMode())
	c.options.Debug = false
	c.SetAuthMode(AuthModePublic)
	assert.Equal(t, AuthModePasswd, c.AuthMode())
	c.options.Debug = true
	c.SetAuthMode(AuthModePublic)
	assert.Equal(t, AuthModePublic, c.AuthMode())
	c.SetAuthMode(AuthModePasswd)
	assert.Equal(t, AuthModePasswd, c.AuthMode())
	c.options.Debug = false
}

func TestAuthSecret(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "", c.AuthSecret())
	c.options.AuthSecret = "341e1657d37759410de1ae628b95dbaa"
	assert.Equal(t, "341e1657d37759410de1ae628b95dbaa", c.AuthSecret())
	c.options.AuthSecret = ""
	assert.Equal(t, "", c.AuthSecret())
}

func TestConfig_AdminPassword(t *testing.T) {
	c := NewConfig(CliTestContext())

	defaultPassword := "photoprism"
	assert.Equal(t, defaultPassword, c.AdminPassword())

	// Test setting the password via secret file.
	_ = os.Setenv(FlagFileVar("ADMIN_PASSWORD"), "testdata/secret_admin")
	assert.Equal(t, defaultPassword, c.AdminPassword())
	c.options.AdminPassword = ""
	assert.Equal(t, "Foo-Bar23", c.AdminPassword())
	_ = os.Setenv(FlagFileVar("ADMIN_PASSWORD"), "")
	c.options.AdminPassword = defaultPassword

	assert.Equal(t, defaultPassword, c.AdminPassword())
}

func TestConfig_AdminScope(t *testing.T) {
	c := NewConfig(CliTestContext())

	// Defaults to empty when no scope was configured.
	assert.Equal(t, "", c.AdminScope())

	// Sanitizes scope attributes using clean.Scope().
	c.options.AdminScope = "  Photos:View   LOGS:* "
	assert.Equal(t, "logs:* photos:view", c.AdminScope())
}

func TestConfig_PasswordLength(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, 8, c.PasswordLength())
	c.options.PasswordLength = 2
	assert.Equal(t, 2, c.PasswordLength())
	c.options.PasswordLength = 30
	assert.Equal(t, 30, c.PasswordLength())
	c.options.PasswordLength = 10000
	assert.Equal(t, 72, c.PasswordLength())
	assert.Equal(t, txt.ClipPassword, c.PasswordLength())
	c.options.PasswordLength = -1
	assert.Equal(t, 8, c.PasswordLength())
	assert.Equal(t, entity.PasswordLengthDefault, c.PasswordLength())
	c.options.PasswordLength = 0
	assert.Equal(t, 8, c.PasswordLength())
}

func TestPasswordResetUri(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "", c.PasswordResetUri())
}

func TestConfig_RegisterUri(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "", c.RegisterUri())
}

func TestConfig_LoginUri(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "/library/login", c.LoginUri())
	c.options.FrontendUri = "/portal/"
	assert.Equal(t, "/portal/login", c.LoginUri())
}

func TestConfig_LoginInfo(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "", c.LoginInfo())
	c.options.LoginInfo = "Foo Bar"
	assert.Equal(t, "Foo Bar", c.LoginInfo())
	c.options.LoginInfo = ""
	assert.Equal(t, "", c.LoginInfo())
}

func TestSessionMaxAge(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, DefaultSessionMaxAge, c.SessionMaxAge())
	c.options.SessionMaxAge = -1
	assert.Equal(t, int64(0), c.SessionMaxAge())
	c.options.SessionMaxAge = 0
	assert.Equal(t, DefaultSessionMaxAge, c.SessionMaxAge())
}

func TestSessionTimeout(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, DefaultSessionTimeout, c.SessionTimeout())
	c.options.SessionTimeout = -1
	assert.Equal(t, int64(0), c.SessionTimeout())
	c.options.SessionTimeout = 0
	assert.Equal(t, DefaultSessionTimeout, c.SessionTimeout())
}

func TestSessionCache(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, DefaultSessionCache, c.SessionCache())
	c.options.SessionCache = -1
	assert.Equal(t, int64(60), c.SessionCache())
	c.options.SessionCache = 100000
	assert.Equal(t, int64(3600), c.SessionCache())
	c.options.SessionCache = 0
	assert.Equal(t, DefaultSessionCache, c.SessionCache())
	assert.Equal(t, time.Duration(DefaultSessionCache)*time.Second, c.SessionCacheDuration())
}

func TestUtils_CheckPassword(t *testing.T) {
	c := NewConfig(CliTestContext())

	formPassword := "photoprism"

	c.options.AdminPassword = "$2b$10$cRhWIleqJkbaFWhBMp54VOI25RvVubxOooCWzWgdrvl5COFxaBnAy"
	check := c.CheckPassword(formPassword)
	assert.True(t, check)

	c.options.AdminPassword = "photoprism"
	check = c.CheckPassword(formPassword)
	assert.True(t, check)

	c.options.AdminPassword = "$2b$10$yprZEQzm/Qy7AaePXtKfkem0kANBZgRwl8HbLE4JrjK6/8Pypgi1W"
	check = c.CheckPassword(formPassword)
	assert.False(t, check)

	c.options.AdminPassword = "admin"
	check = c.CheckPassword(formPassword)
	assert.False(t, check)
}

func TestUtils_isBcrypt(t *testing.T) {
	p := "$2b$10$cRhWIleqJkbaFWhBMp54VOI25RvVubxOooCWzWgdrvl5COFxaBnAy"
	assert.True(t, isBcrypt(p))

	p = "$2b$10$cRhWIleqJkbaFWhBMp54VOI25RvVubxOooCWzWgdrvl5COFxaBnA"
	assert.False(t, isBcrypt(p))

	p = "admin"
	assert.False(t, isBcrypt(p))

	p = ""
	assert.False(t, isBcrypt(p))
}

func TestConfig_KeysPath(t *testing.T) {
	c := NewMinimalTestConfig(t.TempDir())
	assert.Equal(t, filepath.Join(c.ConfigPath(), "keys"), c.KeysPath())
}

func TestConfig_TokenSigningKey(t *testing.T) {
	t.Run("GeneratesStableKeyAtKeysPath", func(t *testing.T) {
		c := NewMinimalTestConfig(t.TempDir())
		key := c.TokenSigningKey()
		assert.GreaterOrEqual(t, len(key), tokens.KeyLen)
		// Stable across calls so tokens stay valid.
		assert.Equal(t, key, c.TokenSigningKey())
		// Persisted at config/keys/signing.key, with no backup copy.
		assert.FileExists(t, filepath.Join(c.KeysPath(), signingKeyName))
		assert.NoFileExists(t, c.BackupPath(signingKeyName))
	})
	t.Run("NonEmptyEvenWhenNotPersisted", func(t *testing.T) {
		c := NewMinimalTestConfig(t.TempDir())
		// Block the keys directory by placing a file where it would be created, so the key cannot be
		// written; it must still be available in memory (never empty).
		require.NoError(t, os.MkdirAll(c.ConfigPath(), fs.ModeDir))
		require.NoError(t, os.WriteFile(c.KeysPath(), []byte("x"), fs.ModeSecretFile))
		key := c.TokenSigningKey()
		assert.GreaterOrEqual(t, len(key), tokens.KeyLen)
		assert.NoFileExists(t, filepath.Join(c.KeysPath(), signingKeyName))
	})
}

func TestConfig_DownloadTokenMaxAge(t *testing.T) {
	c := NewConfig(CliTestContext())
	t.Run("DefaultsToTtlWindow", func(t *testing.T) {
		c.options.DownloadTokenMaxAge = 0
		assert.Equal(t, time.Duration(ttl.DownloadToken.Int())*time.Second, c.DownloadTokenMaxAge())
		// The default must stay well under the session lifetime so a leaked token expires quickly.
		assert.Less(t, int64(c.DownloadTokenMaxAge().Seconds()), c.SessionMaxAge())
	})
	t.Run("ConfiguredOverride", func(t *testing.T) {
		c.options.DownloadTokenMaxAge = 60
		defer func() { c.options.DownloadTokenMaxAge = 0 }()
		assert.Equal(t, 60*time.Second, c.DownloadTokenMaxAge())
	})
}

func TestConfig_InvalidPreviewToken(t *testing.T) {
	c := NewConfig(CliTestContext())

	// See TestConfig_InvalidDownloadToken: pin the shared validation switch so the
	// result is independent of test order.
	validate := entity.ValidateTokens
	defer func() { entity.ValidateTokens = validate }()

	t.Run("ValidationEnabled", func(t *testing.T) {
		entity.ValidateTokens = true
		assert.True(t, c.InvalidPreviewToken("xxx"))
	})
	t.Run("PublicMode", func(t *testing.T) {
		entity.ValidateTokens = false
		assert.False(t, c.InvalidPreviewToken("xxx"))
	})
}
