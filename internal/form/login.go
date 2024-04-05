package form

import (
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Login represents a login form.
type Login struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Code     string `json:"code,omitempty"`
	Token    string `json:"token,omitempty"`
	Email    string `json:"email,omitempty"`
}

// CleanUsername returns the sanitized and normalized username.
func (f Login) CleanUsername() string {
	return clean.Username(f.Username)
}

// CleanEmail returns the sanitized and normalized email.
func (f Login) CleanEmail() string {
	return clean.Email(f.Email)
}

// HasCredentials checks if all credentials is set.
func (f Login) HasCredentials() bool {
	return f.HasUsername() && f.HasPassword()
}

// HasUsername checks if a username is set.
func (f Login) HasUsername() bool {
	if l := len(f.CleanUsername()); l == 0 || l > txt.ClipUsername {
		return false
	}
	return true
}

// HasPassword checks if a password is set.
func (f Login) HasPassword() bool {
	return f.Password != "" && len(f.Password) <= txt.ClipPassword
}

// HasPasscode checks if a verification passcode has been provided.
func (f Login) HasPasscode() bool {
	return clean.Passcode(f.Code) != ""
}

// Passcode returns the sanitized verification passcode.
func (f Login) Passcode() string {
	return clean.Passcode(f.Code)
}

// HasShareToken checks if a link share token has been provided.
func (f Login) HasShareToken() bool {
	return f.Token != ""
}
