package form

import (
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Passcode represents a multi-factor authentication key setup form.
type Passcode struct {
	Type     string `form:"type" json:"type,omitempty"`
	Password string `form:"password" json:"password,omitempty"`
	Code     string `form:"code" json:"code,omitempty"`
}

// HasPassword checks if a password has been provided.
func (f Passcode) HasPassword() bool {
	return f.Password != "" && len(f.Password) <= txt.ClipPassword
}

// HasPasscode checks if a verification code has been provided.
func (f Passcode) HasPasscode() bool {
	return clean.Passcode(f.Code) != ""
}

// Passcode returns the sanitized verification code.
func (f Passcode) Passcode() string {
	return clean.Passcode(f.Code)
}
