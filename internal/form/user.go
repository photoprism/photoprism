package form

import "github.com/photoprism/photoprism/pkg/clean"

// UserCreate represents a User with a new password.
type UserCreate struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UsernameClean returns the username in lowercase and with whitespace trimmed.
func (f UserCreate) UsernameClean() string {
	return clean.Login(f.Username)
}
