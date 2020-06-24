package session

import "github.com/photoprism/photoprism/internal/entity"

type Saved struct {
	UID        string   `json:"uid"`
	Tokens     []string `json:"tokens"`
	Expiration int64    `json:"expiration"`
}

type Data struct {
	User   entity.Person `json:"user"`   // Session user, guest or anonymous person.
	Tokens []string      `json:"tokens"` // Slice of secret share tokens.
	Shared []string      `json:"shared"` // Slice of shared entity UIDs.
}

func (data Data) Saved() Saved {
	return Saved{UID: data.User.PersonUID, Tokens: data.Tokens}
}
