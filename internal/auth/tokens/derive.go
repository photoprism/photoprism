package tokens

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

const (
	// PurposePreview labels the token that authorizes thumbnail and video preview URLs.
	PurposePreview = "preview"
	// deriveLen is the number of HMAC bytes kept, rendered as twice as many hex characters.
	deriveLen = 8
)

// Derive returns a stable token derived from the instance signing key for the given purpose, or "" when
// no usable key or purpose is provided.
// HMAC-SHA256 keeps the key unrecoverable from the token, which is published in URLs, and the purpose
// separates the tokens of different kinds so one cannot be replayed as another.
func Derive(key []byte, purpose string) string {
	if len(key) < KeyLen || purpose == "" {
		return ""
	}

	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(purpose))

	return hex.EncodeToString(mac.Sum(nil)[:deriveLen])
}
