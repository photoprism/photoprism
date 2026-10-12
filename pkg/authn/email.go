package authn

import (
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// VerifiedEmail returns the sanitized address if the identity provider has verified it and it contains
// only ASCII characters as provided, or an empty string otherwise.
func VerifiedEmail(email string, verified bool) string {
	if !verified || !txt.IsASCII(email) {
		return ""
	}

	return clean.Email(email)
}
