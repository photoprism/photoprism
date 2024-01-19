package entity

import (
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// NewClientAccessToken returns a new access token session instance
// that can be used to access the API with unregistered clients.
func NewClientAccessToken(clientName string, lifetime int64, scope string, user *User) *Session {
	sess := NewSession(lifetime, 0)

	if clientName == "" {
		clientName = rnd.Name()
	}

	sess.SetClientName(clientName)
	sess.SetProvider(authn.ProviderAccessToken)
	sess.SetScope(scope)

	if user != nil {
		sess.SetUser(user)
		sess.SetAuthToken(rnd.AuthSecret())
		sess.SetMethod(authn.MethodPersonal)
	} else {
		sess.SetMethod(authn.MethodDefault)
	}

	return sess
}

// CreateClientAccessToken initializes and creates a new access token session
// that can be used to access the API with unregistered clients.
func CreateClientAccessToken(clientName string, lifetime int64, scope string, user *User) (*Session, error) {
	sess := NewClientAccessToken(clientName, lifetime, scope, user)

	if err := sess.Create(); err != nil {
		return nil, err
	}

	return sess, nil
}
