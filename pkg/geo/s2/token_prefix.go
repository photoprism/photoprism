package s2

import (
	"strings"
)

var TokenPrefix = "s2:"

// NormalizeToken removes the prefix from a token and converts all characters to lower case.
func NormalizeToken(token string) string {
	token = strings.ToLower(token)
	token = strings.TrimSpace(token)

	if strings.HasPrefix(token, TokenPrefix) {
		return token[len(TokenPrefix):]
	}

	return token
}

// Prefix adds a token prefix if not exists.
func Prefix(token string) string {
	if len(token) < 3 {
		return token
	}

	if strings.HasPrefix(token, TokenPrefix) {
		return token
	}

	return TokenPrefix + token
}

// PrefixedToken returns the prefixed S2 cell token for coordinates using the default level.
func PrefixedToken(lat, lng float64) string {
	return Prefix(Token(lat, lng))
}
