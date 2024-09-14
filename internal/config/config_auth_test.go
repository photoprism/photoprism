package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
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

func TestPasswordLength(t *testing.T) {
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

func TestRegisterUri(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "", c.RegisterUri())
}

func TestLoginUri(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "/library/login", c.LoginUri())
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

func TestConfig_InvalidDownloadToken(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.True(t, c.InvalidDownloadToken("xxx"))
}

func TestConfig_InvalidPreviewToken(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.True(t, c.InvalidPreviewToken("xxx"))
}
