package entity

import (
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// NewClientSession returns a new session that authenticates a client application.
func NewClientSession(clientName string, expiresIn int64, scope string, grantType authn.GrantType, user *User) *Session {
	sess := NewSession(expiresIn, 0)

	if clientName == "" {
		clientName = rnd.Name()
	}

	sess.SetClientName(clientName)
	sess.SetScope(scope)
	sess.SetGrantType(grantType)

	if user != nil {
		sess.SetUser(user)
		sess.SetAuthToken(rnd.AppPassword())
		sess.SetProvider(authn.ProviderApplication)
		sess.SetMethod(authn.MethodDefault)
	} else {
		sess.SetProvider(authn.ProviderAccessToken)
		sess.SetMethod(authn.MethodOAuth2)
	}

	return sess
}

// AddClientSession creates a new session for authenticating a client application.
func AddClientSession(clientName string, expiresIn int64, scope string, grantType authn.GrantType, user *User) (*Session, error) {
	sess := NewClientSession(clientName, expiresIn, scope, grantType, user)

	if err := sess.Create(); err != nil {
		return nil, err
	}

	return sess, nil
}
