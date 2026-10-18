package oauth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"regexp"
)

// PKCEMethodS256 is the only code challenge method that is accepted.
const PKCEMethodS256 = "S256"

// codeChallengePattern matches a base64url-encoded SHA-256 code challenge (RFC 7636).
var codeChallengePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)

// codeVerifierPattern matches a code verifier as defined in RFC 7636, section 4.1.
var codeVerifierPattern = regexp.MustCompile(`^[A-Za-z0-9._~-]{43,128}$`)

// ValidCodeChallenge reports whether challenge is a well-formed S256 code challenge.
func ValidCodeChallenge(challenge string) bool {
	return codeChallengePattern.MatchString(challenge)
}

// VerifyPKCE reports whether the S256 hash of verifier matches the challenge, comparing in constant time.
func VerifyPKCE(verifier, challenge, method string) bool {
	if method != PKCEMethodS256 || !ValidCodeChallenge(challenge) || !codeVerifierPattern.MatchString(verifier) {
		return false
	}

	sum := sha256.Sum256([]byte(verifier))
	computed := base64.RawURLEncoding.EncodeToString(sum[:])

	return subtle.ConstantTimeCompare([]byte(computed), []byte(challenge)) == 1
}
