package authn

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
)

// PKCE code-challenge methods per RFC 7636 §4.2. Only S256 is accepted; the
// plain method is intentionally not supported because it provides no protection
// when the authorization request is observed.
const (
	PKCEMethodS256  = "S256"
	PKCEMethodPlain = "plain"
)

// ComputePKCEChallenge returns the RFC 7636 §4.2 S256 challenge for verifier,
// BASE64URL(SHA256(ASCII(verifier))) without padding, as stored by the
// authorize endpoint and recomputed at the token endpoint.
func ComputePKCEChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// VerifyPKCE reports whether verifier matches the stored challenge per RFC 7636
// §4.6: challenge == BASE64URL(SHA256(ASCII(verifier))). Only the S256 method is
// accepted; the plain method, any other value, and an empty verifier or
// challenge all return false. The comparison is constant-time so a partial match
// cannot be detected through response timing.
func VerifyPKCE(verifier, challenge, method string) bool {
	if method != PKCEMethodS256 {
		return false
	}
	if verifier == "" || challenge == "" {
		return false
	}
	computed := ComputePKCEChallenge(verifier)
	return subtle.ConstantTimeCompare([]byte(computed), []byte(challenge)) == 1
}
