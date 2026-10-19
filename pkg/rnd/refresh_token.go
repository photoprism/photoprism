package rnd

import (
	"strings"
)

// RefreshToken returns a new refresh token for the grant with the specified UID: the UID, a dot,
// and a random auth token as secret.
func RefreshToken(grantUID string) string {
	return grantUID + "." + AuthToken()
}

// ParseRefreshToken returns the grant UID with the specified prefix and the secret of a refresh token,
// and false if the token does not have this format.
func ParseRefreshToken(s string, prefix byte) (grantUID, secret string, ok bool) {
	grantUID, secret, found := strings.Cut(s, ".")

	if !found || !IsUID(grantUID, prefix) || !IsAuthToken(secret) {
		return "", "", false
	}

	return grantUID, secret, true
}
