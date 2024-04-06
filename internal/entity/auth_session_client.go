package entity

import (
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// NewClientAuthentication returns a new session that authenticates a client application.
func NewClientAuthentication(clientName string, lifetime int64, scope string, grantType authn.GrantType, user *User) *Session {
	sess := NewSession(lifetime, 0)

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

// AddClientAuthentication creates a new session for authenticating a client application.
func AddClientAuthentication(clientName string, lifetime int64, scope string, grantType authn.GrantType, user *User) (*Session, error) {
	sess := NewClientAuthentication(clientName, lifetime, scope, grantType, user)

	if err := sess.Create(); err != nil {
		return nil, err
	}

	return sess, nil
}
