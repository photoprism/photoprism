package tokens

import (
	"net/url"
	"time"

	"github.com/photoprism/photoprism/pkg/authn/authtoken"
)

const (
	// KeyLen is the minimum byte length of the HMAC keys that sign URL tokens.
	KeyLen = 32
	// maxTokenLen bounds the length of a token accepted for verification, so a malformed request
	// cannot force an oversized constant-time comparison.
	maxTokenLen = 512
)

// Signer holds the configured signing material for one kind of bunny.net-compatible URL token.
// config.Propagate sets Key and SignaturePath on each package-level Signer (see Download), so this
// leaf never imports config or get. The zero value is unconfigured and refuses to mint or verify, so
// a missing key degrades to rejection rather than to a forgeable empty-key HMAC.
type Signer struct {
	Key           []byte // HMAC key; an instance secret that never reaches client config.
	SignaturePath string // URL path prefix the signature is bound to (a directory-style token).
}

// Configured reports whether the signer has a usable key.
func (s *Signer) Configured() bool {
	return s != nil && len(s.Key) >= KeyLen
}

// Sign returns the HS256- token authorizing a request with the given expiry and signed parameters,
// or "" when the signer is unconfigured. Expiry and the parameter set are chosen by the caller so a
// kind's own policy (sliding vs. bucketed expiry, per-session vs. per-path scope) stays out of the
// generic signer.
func (s *Signer) Sign(expires int64, params url.Values) string {
	if !s.Configured() {
		return ""
	}

	return authtoken.Sign(s.Key, s.SignaturePath, expires, params, "")
}

// Valid reports whether token is an unexpired, correctly-signed value for the given expiry and
// parameters. It returns false for an unconfigured signer or an empty/oversized token.
func (s *Signer) Valid(expires int64, params url.Values, token string) bool {
	if !s.Configured() || token == "" || len(token) > maxTokenLen {
		return false
	}

	return authtoken.Valid(s.Key, s.SignaturePath, expires, params, "", token, time.Now().Unix())
}
