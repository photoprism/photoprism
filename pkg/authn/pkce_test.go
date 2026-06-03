package authn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// RFC 7636 Appendix B test vector.
const (
	rfcPKCEVerifier  = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	rfcPKCEChallenge = "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
)

func TestComputePKCEChallenge(t *testing.T) {
	t.Run("RFCVector", func(t *testing.T) {
		assert.Equal(t, rfcPKCEChallenge, ComputePKCEChallenge(rfcPKCEVerifier))
	})
	t.Run("Empty", func(t *testing.T) {
		// SHA-256 of the empty string, base64url without padding.
		assert.Equal(t, "47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU", ComputePKCEChallenge(""))
	})
}

func TestVerifyPKCE(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.True(t, VerifyPKCE(rfcPKCEVerifier, rfcPKCEChallenge, PKCEMethodS256))
	})
	t.Run("WrongVerifier", func(t *testing.T) {
		assert.False(t, VerifyPKCE("not-the-verifier", rfcPKCEChallenge, PKCEMethodS256))
	})
	t.Run("WrongChallenge", func(t *testing.T) {
		assert.False(t, VerifyPKCE(rfcPKCEVerifier, "wrongchallenge", PKCEMethodS256))
	})
	t.Run("PlainRejected", func(t *testing.T) {
		assert.False(t, VerifyPKCE(rfcPKCEVerifier, rfcPKCEVerifier, PKCEMethodPlain))
	})
	t.Run("EmptyMethod", func(t *testing.T) {
		assert.False(t, VerifyPKCE(rfcPKCEVerifier, rfcPKCEChallenge, ""))
	})
	t.Run("EmptyVerifier", func(t *testing.T) {
		assert.False(t, VerifyPKCE("", rfcPKCEChallenge, PKCEMethodS256))
	})
	t.Run("EmptyChallenge", func(t *testing.T) {
		assert.False(t, VerifyPKCE(rfcPKCEVerifier, "", PKCEMethodS256))
	})
}
