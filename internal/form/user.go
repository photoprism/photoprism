package form

import "github.com/photoprism/photoprism/pkg/sanitize"

// UserCreate represents a User with a new password.
type UserCreate struct {
	UserName string `json:"username"`
	FullName string `json:"fullname"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Username returns the normalized username in lowercase and without whitespace padding.
func (f UserCreate) Username() string {
	return sanitize.Username(f.UserName)
}
