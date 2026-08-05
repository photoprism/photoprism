package authtoken

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGoldenVector pins the token wire format to a frozen value for fixed inputs and cross-checks it
// against a from-scratch HMAC computation written to the documented bunny.net Advanced algorithm
// (token = "HS256-" + base64url(HMAC_SHA256(key, signaturePath + expires + signingData + clientIP))).
// A change to the message layout, SigningData joining, or base64 encoding breaks this test before it
// can silently diverge from what a bunny.net edge would validate. Regenerate the literal only when the
// format intentionally changes, and verify the new value against the bunny.net Go reference (go/token.go
// at github.com/BunnyWay/BunnyCDN.TokenAuthentication).
func TestGoldenVector(t *testing.T) {
	var (
		key      = []byte("0123456789abcdef0123456789abcdef") // fixed 32-byte key
		path     = "/api/v1"
		expires  = int64(1_700_000_000)
		params   = url.Values{"sid": {"sess123"}}
		clientIP = ""
		// golden: computed independently as HS256- + base64url(HMAC_SHA256(key, "/api/v11700000000sid=sess123")).
		golden = "HS256-rhHSpkbnhDCUP3jATEHp51DluVpFjfMeuTADYS2KlSA"
	)

	t.Run("SignMatchesGolden", func(t *testing.T) {
		assert.Equal(t, golden, Sign(key, path, expires, params, clientIP))
	})
	t.Run("IndependentRecomputeMatchesGolden", func(t *testing.T) {
		// Rebuild the message and MAC from scratch, per the documented formula, without using Sign.
		mac := hmac.New(sha256.New, key)
		mac.Write([]byte(path + "1700000000" + "sid=sess123" + clientIP))
		want := Prefix + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		assert.Equal(t, golden, want)
	})
	t.Run("VerifyAcceptsGolden", func(t *testing.T) {
		assert.NoError(t, Verify(key, path, expires, params, clientIP, golden, expires-1))
	})
}
