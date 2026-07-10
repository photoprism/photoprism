package oidc

import (
	"net/url"
)

// EndSessionURL builds the RP-initiated logout URL for the provider's end_session_endpoint,
// or "" (no error) when none is advertised so callers fall back to a local-only logout.
// The result must be followed by a browser navigation, not a server-side request (only the
// browser carries the provider SSO cookie) — hence building the URL rather than rp.EndSession.
func (c *Client) EndSessionURL(idToken, postLogoutRedirectURI, state string) (string, error) {
	endpoint := c.GetEndSessionEndpoint()

	if endpoint == "" {
		return "", nil
	}

	u, err := url.Parse(endpoint)

	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("id_token_hint", idToken)

	if clientID := c.OAuthConfig().ClientID; clientID != "" {
		q.Set("client_id", clientID)
	}

	if postLogoutRedirectURI != "" {
		q.Set("post_logout_redirect_uri", postLogoutRedirectURI)
	}

	if state != "" {
		q.Set("state", state)
	}

	u.RawQuery = q.Encode()

	return u.String(), nil
}
