package oauth

import "github.com/photoprism/photoprism/pkg/rnd"

// HashSecret returns the hex-encoded SHA-256 hash under which a code, request ID or refresh token is stored.
func HashSecret(raw string) string {
	return rnd.Sha256([]byte(raw))
}
