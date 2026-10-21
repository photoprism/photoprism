package oauth

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"time"

	"github.com/photoprism/photoprism/internal/auth/tokens"
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

	grantUID := rnd.GenerateUID(entity.GrantUID)
	refreshToken, err := newRefreshToken(grantUID)

	if err != nil {
		return nil, "", err
	}

	// A user has at most one grant per app; approving it again replaces the previous one and its tokens.
	if prev := entity.FindUserClientGrant(user.UserUID, code.ClientID); prev != nil {
		if err = prev.Delete(); err != nil {
			return nil, "", err
		}
	}

	m := &entity.Grant{
		GrantUID:    grantUID,
		UserUID:     user.UserUID,
		UserName:    user.Username(),
		ClientID:    code.ClientID,
		ClientName:  clientName,
		RedirectURI: code.RedirectURI,
		Scope:       code.Scope,
		Resource:    code.Resource,
	}

	m.RefreshHash = HashSecret(refreshToken)

	if err = m.Create(); err != nil {
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

// FindRefreshGrant returns the grant a signed refresh token names, without consuming the token. A token
// without a valid signature is refused before the grant is looked up.
func FindRefreshGrant(raw string) *entity.Grant {
	if grantUID := refreshGrantUID(raw); grantUID != "" {
		return entity.FindGrant(grantUID)
	}

	return nil
}

// RefreshGrant rotates the refresh token of the grant found with FindRefreshGrant and returns the grant with
// the new raw refresh token. A refresh token that does not match the current one revokes the grant, as does a
// concurrent rotation.
func RefreshGrant(m *entity.Grant, raw, clientID string) (*entity.Grant, string, error) {
	if m == nil || refreshGrantUID(raw) != m.GrantUID || clientID == "" || m.ClientID != clientID {
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

	next, err := rotateGrant(m, hash)

	if errors.Is(err, ErrRefreshTokenReused) {
		return m, "", err
	} else if err != nil {
		return nil, "", err
	}

	return m, next, nil
}

// rotateGrant replaces the refresh token of a grant whose current hash is the specified one and returns
// the new raw token. If another request rotated it first, the grant is revoked as for a reused token.
func rotateGrant(m *entity.Grant, hash string) (string, error) {
	next, err := newRefreshToken(m.GrantUID)

	if err != nil {
		return "", err
	}

	if changed, rotateErr := m.Rotate(hash, HashSecret(next)); rotateErr != nil {
		return "", rotateErr
	} else if !changed {
		if err = m.Delete(); err != nil {
			return "", err
		}

		return "", ErrRefreshTokenReused
	}

	return next, nil
}

// RevokeRefreshToken deletes the grant of a refresh token and its access tokens if the token is current and,
// if a client ID is specified, belongs to this client, and returns the revoked grant.
func RevokeRefreshToken(raw, clientID string) (*entity.Grant, error) {
	m := FindRefreshGrant(raw)

	if m == nil || clientID != "" && clientID != m.ClientID || subtle.ConstantTimeCompare([]byte(HashSecret(raw)), []byte(m.RefreshHash)) != 1 {
		return nil, nil
	}

	if err := m.Delete(); err != nil {
		return nil, err
	}

	return m, nil
}

// newRefreshToken returns a new signed refresh token for the grant.
func newRefreshToken(grantUID string) (string, error) {
	if raw := tokens.SignRefresh(rnd.RefreshToken(grantUID)); raw != "" {
		return raw, nil
	}

	return "", ErrSigningKey
}

// refreshGrantUID returns the grant UID of a signed refresh token, or "" if the token or its signature is invalid.
func refreshGrantUID(raw string) string {
	value, ok := tokens.VerifyRefresh(raw)

	if !ok {
		return ""
	}

	grantUID, _, ok := rnd.ParseRefreshToken(value, entity.GrantUID)

	if !ok {
		return ""
	}

	return grantUID
}

// PurgeIdleGrants deletes grants that have not been used for GrantMaxIdle.
func PurgeIdleGrants() {
	if n := entity.DeleteIdleGrants(time.Now().UTC().Add(-GrantMaxIdle)); n > 0 {
		log.Debugf("oauth: deleted %d idle grants", n)
	}
}
