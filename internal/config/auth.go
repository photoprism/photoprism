package config

import (
	"regexp"

	"github.com/photoprism/photoprism/pkg/rnd"
	"golang.org/x/crypto/bcrypt"
)

func isBcrypt(s string) bool {
	b, err := regexp.MatchString(`^\$2[ayb]\$.{56}$`, s)
	if err != nil {
		return false
	}
	return b
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

// InvalidDownloadToken returns true if the token is invalid.
func (c *Config) InvalidDownloadToken(t string) bool {
	return c.DownloadToken() != t
}

// DownloadToken returns the DOWNLOAD api token (you can optionally use a static value for permanent caching).
func (c *Config) DownloadToken() string {
	if c.params.DownloadToken == "" {
		c.params.DownloadToken = rnd.Token(8)
	}

	return c.params.DownloadToken
}

// InvalidToken returns true if the token is invalid.
func (c *Config) InvalidToken(t string) bool {
	return c.ThumbToken() != t && c.DownloadToken() != t
}

// ThumbToken returns the THUMBNAILS api token (you can optionally use a static value for permanent caching).
func (c *Config) ThumbToken() string {
	if c.params.ThumbToken == "" {
		c.params.ThumbToken = rnd.Token(8)
	}

	return c.params.ThumbToken
}
