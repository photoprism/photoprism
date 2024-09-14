package authn

import "github.com/photoprism/photoprism/pkg/clean"

// IssuerUri represents an authentication issuer URI.
type IssuerUri = string

const (
	IssuerDefault IssuerUri = ""
)

// Issuer returns a sanitized issuer URI.
func Issuer(uri string) IssuerUri {
	return clean.Uri(uri)
}
