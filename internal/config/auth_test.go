package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
