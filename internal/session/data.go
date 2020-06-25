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

type UIDs []string

func (list UIDs) String() string {
	return strings.Join(list, ",")
}

type Data struct {
	User   entity.Person `json:"user"`   // Session user, guest or anonymous person.
	Tokens []string      `json:"tokens"` // Slice of secret share tokens.
	Shares UIDs          `json:"shares"` // Slice of shared entity UIDs.
}

func (s Data) Saved() Saved {
	return Saved{User: s.User.PersonUID, Tokens: s.Tokens}
}

func (s Data) Invalid() bool {
	return s.User.ID == 0 || s.User.PersonUID == "" || (s.Guest() && s.NoShares())
}

func (s Data) Valid() bool {
	return !s.Invalid()
}

func (s Data) Guest() bool {
	return s.User.Guest()
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
