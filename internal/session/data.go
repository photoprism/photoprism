package session

import (
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
)

type Saved struct {
	User       string   `json:"user"`
	Tokens     []string `json:"tokens"`
	Expiration int64    `json:"expiration"`
}

// UIDs represents a slice of unique ID strings.
type UIDs []string

// String returns all UIDs as comma separated string.
func (u UIDs) String() string {
	return u.Join(",")
}

// Join returns all UIDs as custom separated string.
func (u UIDs) Join(s string) string {
	return strings.Join(u, s)
}

type Data struct {
	User   entity.User `json:"user"`   // Session user, guest or anonymous person.
	Tokens []string    `json:"tokens"` // Slice of secret share tokens.
	Shares UIDs        `json:"shares"` // Slice of shared entity UIDs.
}

func (s Data) Saved() Saved {
	return Saved{User: s.User.UserUID, Tokens: s.Tokens}
}

func (s Data) Invalid() bool {
	return s.User.ID == 0 || s.User.UserUID == "" || (s.Guest() && s.NoShares())
}

func (s Data) Valid() bool {
	return !s.Invalid()
}

func (s Data) Guest() bool {
	return s.User.IsGuest()
}

func (s Data) NoShares() bool {
	return len(s.Shares) == 0
}

func (s Data) HasShare(uid string) bool {
	for _, share := range s.Shares {
		if share == uid {
			return true
		}
	}

	return false
}
