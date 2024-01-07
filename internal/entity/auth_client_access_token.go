package entity

import (
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
)

// NewClientAccessToken returns a new access token session instance
// that can be used to access the API with unregistered clients.
func NewClientAccessToken(id string, lifetime int64, scope string, user *User) *Session {
	sess := NewSession(lifetime, 0)

	if id == "" {
		id = TimeStamp().UTC().Format("2006-01-02 15:04:05")
	}

	sess.AuthID = clean.Name(id)
	sess.AuthProvider = authn.ProviderClient.String()
	sess.AuthMethod = authn.MethodAccessToken.String()
	sess.AuthScope = clean.Scope(scope)

	if user != nil {
		sess.SetUser(user)
	}

	return sess
}

// CreateClientAccessToken initializes and creates a new access token session
// that can be used to access the API with unregistered clients.
func CreateClientAccessToken(id string, lifetime int64, scope string, user *User) (*Session, error) {
	sess := NewClientAccessToken(id, lifetime, scope, user)

	if err := sess.Create(); err != nil {
		return nil, err
	}

	return sess, nil
}
