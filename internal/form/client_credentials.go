package form

import (
	"fmt"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// ClientCredentials represents client authentication request credentials.
type ClientCredentials struct {
	ClientID     string `form:"client_id" json:"client_id,omitempty"`
	ClientSecret string `form:"client_secret" json:" client_secret,omitempty"`
	AuthScope    string `form:"scope" json:"scope,omitempty"`
}

// Validate checks the grant type and credentials.
func (f ClientCredentials) Validate() error {
	// Check client ID.
	if f.ClientID == "" {
		return fmt.Errorf("missing client id")
	} else if rnd.InvalidUID(f.ClientID, 'c') {
		return fmt.Errorf("invalid client id")
	}

	// Check client secret.
	if f.ClientSecret == "" {
		return fmt.Errorf("missing client secret")
	} else if !rnd.IsAlnum(f.ClientSecret) {
		return fmt.Errorf("invalid client secret")
	}

	return nil
}

// Scope returns the client scopes as sanitized string.
func (f ClientCredentials) Scope() string {
	return clean.Scope(f.AuthScope)
}
