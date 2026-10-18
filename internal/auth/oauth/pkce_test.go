package oauth

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

// testChallenge returns the S256 code challenge for a verifier.
func testChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

const testVerifier = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"

func TestValidCodeChallenge(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.True(t, ValidCodeChallenge("E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"))
	})
	t.Run("TooShort", func(t *testing.T) {
		assert.False(t, ValidCodeChallenge("E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-c"))
	})
	t.Run("InvalidCharacters", func(t *testing.T) {
		assert.False(t, ValidCodeChallenge("E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw+cM"))
	})
}

func TestVerifyPKCE(t *testing.T) {
	challenge := testChallenge(testVerifier)

	t.Run("RFC7636Example", func(t *testing.T) {
		assert.Equal(t, "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM", challenge)
		assert.True(t, VerifyPKCE(testVerifier, challenge, PKCEMethodS256))
	})
	t.Run("WrongVerifier", func(t *testing.T) {
		assert.False(t, VerifyPKCE(testVerifier+"x", challenge, PKCEMethodS256))
	})
	t.Run("PlainMethod", func(t *testing.T) {
		assert.False(t, VerifyPKCE(testVerifier, testVerifier, "plain"))
	})
	t.Run("ShortVerifier", func(t *testing.T) {
		v := "abc"
		assert.False(t, VerifyPKCE(v, testChallenge(v), PKCEMethodS256))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, VerifyPKCE("", challenge, PKCEMethodS256))
		assert.False(t, VerifyPKCE(testVerifier, "", PKCEMethodS256))
	})
}
