package oauth

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
)

// ClientChange names the settings a client change request sets explicitly, since an empty form value
// keeps the current one.
type ClientChange struct {
	Method    bool // Method is true if the request sets the authentication method, also when it is empty.
	Role      bool // Role is true if the request sets the client role.
	NewSecret bool // NewSecret is true if the request asks for a new random secret.
}

// ValidateClientForm checks the values of a new client: an app client has redirect URIs, belongs to
// no user and does not use a client secret, and only app clients have redirect URIs.
func ValidateClientForm(frm form.Client) error {
	if !frm.IsApp() {
		if len(frm.RedirectURIs) > 0 {
			return fmt.Errorf("redirect URIs can only be set for app clients")
		}

		return nil
	}

	switch {
	case len(frm.RedirectURIs) == 0:
		return fmt.Errorf("app clients require at least one redirect URI")
	case frm.UserUID != "" || frm.UserName != "":
		return fmt.Errorf("app clients cannot belong to a user")
	case frm.Secret() != "":
		return fmt.Errorf("app clients do not use a client secret")
	case frm.Role() != "" && frm.Role() != acl.RoleClient.String():
		return fmt.Errorf("app clients must have the %s role", acl.RoleClient.String())
	}

	return ValidateRedirectURIs(frm.RedirectURIs)
}

// ValidateClientChange checks changes to an existing client: its auth method cannot be switched to or from
// the app type, app clients keep the client role and no secret, and only app clients have redirect URIs.
func ValidateClientChange(client *entity.Client, frm form.Client, set ClientChange) error {
	if client == nil {
		return fmt.Errorf("client not found")
	}

	isApp := client.IsApp()

	switch {
	case (set.Method || frm.AuthMethod != "") && frm.IsApp() != isApp:
		return fmt.Errorf("the authentication method of app clients cannot be changed")
	case frm.RedirectURIs != nil && !isApp:
		return fmt.Errorf("redirect URIs can only be set for app clients")
	case !isApp:
		return nil
	case frm.Secret() != "" || set.NewSecret:
		return fmt.Errorf("app clients do not use a client secret")
	case set.Role && clean.Role(frm.ClientRole) != acl.RoleClient.String():
		return fmt.Errorf("app clients must have the %s role", acl.RoleClient.String())
	}

	return ValidateRedirectURIs(frm.RedirectURIs)
}

// ValidateRedirectURIs checks that each redirect URI is an HTTPS URL, an HTTP URL on a loopback host,
// or a private-use URI scheme in reverse domain notation.
func ValidateRedirectURIs(uris []string) error {
	for _, uri := range uris {
		if !ValidRedirectURI(uri) || !ValidValue(uri, MaxRedirectURILength) {
			return fmt.Errorf("invalid redirect URI %s", clean.LogQuote(uri))
		}
	}

	return nil
}
