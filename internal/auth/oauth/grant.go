package oauth

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// AccessTokenTTL is the lifetime of an app access token.
const AccessTokenTTL = time.Hour

// GrantMaxIdle is how long a grant remains valid without being used.
const GrantMaxIdle = 90 * 24 * time.Hour

// ErrRefreshTokenReused is returned when a refresh token that has already been used is presented again.
var ErrRefreshTokenReused = fmt.Errorf("%w: refresh token reused", ErrInvalidGrant)

// NewGrant replaces the user's grant for the client of a redeemed authorization code with a new one and
// returns it with its raw refresh token.
func NewGrant(code *entity.AuthCode, user *entity.User, clientName string) (*entity.Grant, string, error) {
	if code == nil || user == nil || user.UserUID != code.UserUID || code.ClientID == "" {
		return nil, "", errors.New("oauth: invalid authorization code")
	}

	// A user has at most one grant per app; approving it again replaces the previous one and its tokens.
	if prev := entity.FindUserClientGrant(user.UserUID, code.ClientID); prev != nil {
		if err := prev.Delete(); err != nil {
			return nil, "", err
		}
	}

	m := &entity.Grant{
		GrantUID:    rnd.GenerateUID(entity.GrantUID),
		UserUID:     user.UserUID,
		UserName:    user.Username(),
		ClientID:    code.ClientID,
		ClientName:  clientName,
		RedirectURI: code.RedirectURI,
		Scope:       code.Scope,
		Resource:    code.Resource,
	}

	refreshToken := rnd.RefreshToken(m.GrantUID)
	m.RefreshHash = HashSecret(refreshToken)

	if err := m.Create(); err != nil {
		return nil, "", err
	}

	return m, refreshToken, nil
}

// NewAccessToken returns a new, unsaved access token session for the grant, valid for AccessTokenTTL and
// accepted only by the grant's resource.
func NewAccessToken(grant *entity.Grant, user *entity.User, grantType authn.GrantType) *entity.Session {
	sess := entity.NewSession(int64(AccessTokenTTL/time.Second), 0)
	sess.SetUser(user)
	sess.SetClientName(grant.ClientName)
	sess.SetProvider(authn.ProviderOAuthApp)
	sess.SetMethod(authn.MethodPKCE)
	sess.SetGrantType(grantType)
	sess.AuthID = grant.GrantUID
	sess.AuthScope = grant.Scope
	sess.AuthAudience = grant.Resource

	return sess
}

// FindRefreshGrant returns the grant a refresh token names, without verifying or consuming the token.
func FindRefreshGrant(raw string) *entity.Grant {
	grantUID, _, ok := rnd.ParseRefreshToken(raw, entity.GrantUID)

	if !ok {
		return nil
	}

	return entity.FindGrant(grantUID)
}

// RefreshGrant rotates the refresh token of a grant and returns the grant with the new raw refresh token.
// A refresh token that does not match the current one revokes the grant, as does a concurrent rotation.
func RefreshGrant(raw, clientID string) (*entity.Grant, string, error) {
	grantUID, _, ok := rnd.ParseRefreshToken(raw, entity.GrantUID)

	if !ok {
		return nil, "", ErrInvalidGrant
	}

	m := entity.FindGrant(grantUID)

	if m == nil || clientID == "" || m.ClientID != clientID {
		return nil, "", ErrInvalidGrant
	}

	// Grants end after a period of inactivity.
	if m.Idle(time.Now().UTC().Add(-GrantMaxIdle)) {
		if err := m.Delete(); err != nil {
			return nil, "", err
		}

		return nil, "", ErrInvalidGrant
	}

	hash := HashSecret(raw)

	// Revoke the grant if a refresh token is used again.
	if subtle.ConstantTimeCompare([]byte(hash), []byte(m.RefreshHash)) != 1 {
		if err := m.Delete(); err != nil {
			return nil, "", err
		}

		return m, "", ErrRefreshTokenReused
	}

	next := rnd.RefreshToken(m.GrantUID)

	if changed, err := m.Rotate(hash, HashSecret(next)); err != nil {
		return nil, "", err
	} else if !changed {
		if err = m.Delete(); err != nil {
			return nil, "", err
		}

		return m, "", ErrRefreshTokenReused
	}

	return m, next, nil
}

// RevokeRefreshToken deletes the grant of a refresh token and its access tokens if the token is current and,
// if a client ID is specified, belongs to this client, and returns the revoked grant.
func RevokeRefreshToken(raw, clientID string) (*entity.Grant, error) {
	grantUID, _, ok := rnd.ParseRefreshToken(raw, entity.GrantUID)

	if !ok {
		return nil, nil
	}

	m := entity.FindGrant(grantUID)

	if m == nil || clientID != "" && clientID != m.ClientID || subtle.ConstantTimeCompare([]byte(HashSecret(raw)), []byte(m.RefreshHash)) != 1 {
		return nil, nil
	}

	if err := m.Delete(); err != nil {
		return nil, err
	}

	return m, nil
}

// PurgeIdleGrants deletes grants that have not been used for GrantMaxIdle.
func PurgeIdleGrants() {
	if n := entity.DeleteIdleGrants(time.Now().UTC().Add(-GrantMaxIdle)); n > 0 {
		log.Debugf("oauth: deleted %d idle grants", n)
	}
}
