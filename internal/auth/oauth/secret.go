package oauth

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashSecret returns the hex-encoded SHA-256 hash under which a code, request ID or refresh token is stored.
func HashSecret(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
