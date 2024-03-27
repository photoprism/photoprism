package form

import (
	"github.com/photoprism/photoprism/pkg/clean"
)

// Login represents a login form.
type Login struct {
	UserName   string `json:"username,omitempty"`
	UserEmail  string `json:"email,omitempty"`
	Password   string `json:"password,omitempty"`
	Passcode   string `json:"passcode,omitempty"`
	ShareToken string `json:"token,omitempty"`
}

// Username returns the sanitized username in lowercase.
func (f Login) Username() string {
	return clean.Username(f.UserName)
}

// Email returns the sanitized email in lowercase.
func (f Login) Email() string {
	return clean.Email(f.UserEmail)
}

// HasUsername checks if a username is set.
func (f Login) HasUsername() bool {
	if l := len(f.Username()); l == 0 || l > 255 {
		return false
	}
	return true
}

// HasPasscode checks if a passcode is set.
func (f Login) HasPasscode() bool {
	return f.Passcode != "" && len(f.Passcode) <= 255
}

// HasPassword checks if a password is set.
func (f Login) HasPassword() bool {
	return f.Password != "" && len(f.Password) <= 255
}

// HasShareToken checks if a link share token has been provided.
func (f Login) HasShareToken() bool {
	return f.ShareToken != ""
}

// HasCredentials checks if all credentials is set.
func (f Login) HasCredentials() bool {
	return f.HasUsername() && f.HasPassword()
}
