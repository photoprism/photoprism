package entity

import (
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// NewClientAccessToken returns a new access token session instance
// that can be used to access the API with unregistered clients.
func NewClientAccessToken(name string, lifetime int64, scope string, user *User) *Session {
	sess := NewSession(lifetime, 0)

	if name == "" {
		name = rnd.Name()
	}

	sess.AuthID = clean.Name(name)
	sess.AuthProvider = authn.ProviderClient.String()
	sess.AuthMethod = authn.MethodAccessToken.String()
	sess.AuthScope = clean.Scope(scope)

	if user != nil {
		sess.SetUser(user)
		sess.SetAuthToken(rnd.AuthSecret())
	}

	return sess
}

// CreateClientAccessToken initializes and creates a new access token session
// that can be used to access the API with unregistered clients.
func CreateClientAccessToken(name string, lifetime int64, scope string, user *User) (*Session, error) {
	sess := NewClientAccessToken(name, lifetime, scope, user)

	if err := sess.Create(); err != nil {
		return nil, err
	}

	return sess, nil
}
