package oidc

import (
	"encoding/base64"
	"errors"

	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/photoprism/photoprism/pkg/rnd"
)

// NonceCookie is the name of the signed and encrypted cookie that transports the
// per-request OIDC nonce from the authorization redirect to the callback.
const NonceCookie = "nonce"

// nonceParam is the OIDC authorization request parameter that carries the nonce.
const nonceParam = "nonce"

// nonceBytes is the number of random bytes used to generate a nonce.
const nonceBytes = 32

// ErrNonceMismatch is returned when an ID token echoes a nonce that does not
// match the value sent on the authorization request.
var ErrNonceMismatch = errors.New("id token nonce does not match")

// Nonce returns a new random nonce for a single OIDC authorization request,
// base64url-encoded so it is safe to use as a URL query parameter.
func Nonce() (string, error) {
	b, err := rnd.RandomBytes(nonceBytes)

	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

// CheckNonce validates the ID token's nonce claim against the value sent for the
// authorization request. It tolerates a provider that omits the nonce on a
// session-resumed token (empty claim) so logins do not regress, but rejects a
// non-empty claim that does not match the expected value.
func CheckNonce(expected string, tokens *oidc.Tokens[*oidc.IDTokenClaims]) error {
	if tokens == nil || tokens.IDTokenClaims == nil {
		return nil
	}

	got := tokens.IDTokenClaims.GetNonce()

	if got == "" {
		return nil
	}

	if got != expected {
		return ErrNonceMismatch
	}

	return nil
}
