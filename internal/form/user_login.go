package form

import (
	"github.com/photoprism/photoprism/pkg/clean"
)

// Login represents a login form.
type Login struct {
	UserName  string `json:"username,omitempty"`
	UserEmail string `json:"email,omitempty"`
	Password  string `json:"password,omitempty"`
	AuthToken string `json:"token,omitempty"`
}

// Name returns the sanitized username in lowercase.
func (f Login) Name() string {
	return clean.DN(f.UserName)
}

// Email returns the sanitized email in lowercase.
func (f Login) Email() string {
	return clean.Email(f.UserEmail)
}

// HasName checks if a username is set.
func (f Login) HasName() bool {
	if l := len(f.Name()); l == 0 || l > 255 {
		return false
	}
	return true
}

// HasPassword checks if a password is set.
func (f Login) HasPassword() bool {
	return f.Password != "" && len(f.Password) <= 255
}

// HasToken checks if an auth token is set.
func (f Login) HasToken() bool {
	return f.AuthToken != ""
}

// HasCredentials checks if all credentials is set.
func (f Login) HasCredentials() bool {
	return f.HasName() && f.HasPassword()
}
