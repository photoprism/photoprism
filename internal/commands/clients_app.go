package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/auth/oauth"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
)

// validateAppClientForm checks the values of a new app client, which has callback URLs, belongs to
// no user and does not use a client secret.
func validateAppClientForm(frm form.Client) error {
	if !frm.IsApp() {
		if len(frm.RedirectURIs) > 0 {
			return fmt.Errorf("callback URLs can only be set for app clients")
		}

		return nil
	}

	switch {
	case len(frm.RedirectURIs) == 0:
		return fmt.Errorf("app clients require at least one callback URL")
	case frm.UserUID != "" || frm.UserName != "":
		return fmt.Errorf("app clients cannot belong to a user")
	case frm.Secret() != "":
		return fmt.Errorf("app clients do not use a client secret")
	case frm.Role() != "" && frm.Role() != acl.RoleClient.String():
		return fmt.Errorf("app clients must have the %s role", acl.RoleClient.String())
	}

	return validateCallbackURLs(frm.RedirectURIs)
}

// validateAppClientMod checks changes to an existing client: its auth method cannot be switched to or from
// the app type, app clients keep the client role and no secret, and only app clients have callback URLs.
func validateAppClientMod(ctx *cli.Context, client *entity.Client, frm form.Client) error {
	isApp := client.IsApp()

	switch {
	case ctx.IsSet("method") && frm.IsApp() != isApp:
		return fmt.Errorf("the authentication method of app clients cannot be changed")
	case frm.RedirectURIs != nil && !isApp:
		return fmt.Errorf("callback URLs can only be set for app clients")
	case !isApp:
		return nil
	case frm.Secret() != "" || ctx.Bool("regenerate"):
		return fmt.Errorf("app clients do not use a client secret")
	case ctx.IsSet("role") && clean.Role(frm.ClientRole) != acl.RoleClient.String():
		return fmt.Errorf("app clients must have the %s role", acl.RoleClient.String())
	}

	return validateCallbackURLs(frm.RedirectURIs)
}

// validateCallbackURLs checks that each callback URL is an HTTPS URL, an HTTP URL on a loopback host,
// or a private-use URI scheme in reverse domain notation.
func validateCallbackURLs(uris []string) error {
	for _, uri := range uris {
		if !oauth.ValidRedirectURI(uri) || !oauth.ValidValue(uri, oauth.MaxRedirectURILength) {
			return fmt.Errorf("invalid callback URL %s", clean.LogQuote(uri))
		}
	}

	return nil
}
