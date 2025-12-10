package mock

// revive:disable
// Dummy storage implementation for the test OIDC provider; lint strictness is relaxed intentionally.

import (
	"time"

	"github.com/zitadel/oidc/pkg/oidc"
	"github.com/zitadel/oidc/pkg/op"
)

// ConfClient represents a fixed client configuration used by the dummy provider.
type ConfClient struct {
	applicationType op.ApplicationType
	authMethod      oidc.AuthMethod
	responseTypes   []oidc.ResponseType
	grantTypes      []oidc.GrantType
	ID              string
	accessTokenType op.AccessTokenType
	devMode         bool
}

func (c *ConfClient) GetID() string {
	return c.ID
}

func (c *ConfClient) RedirectURIs() []string {
	return []string{
		"https://registered.com/callback",
		"http://localhost:9999/callback",
		"http://localhost:5556/auth/callback",
		"custom://callback",
		"https://localhost:8443/test/a/instructions-example/callback",
		"https://op.certification.openid.net:62064/authz_cb",
		"https://op.certification.openid.net:62064/authz_post",
		"http://localhost:2342/api/v1/oidc/redirect",
		"https://app.localssl.dev/api/v1/oidc/redirect",
	}
}
func (c *ConfClient) PostLogoutRedirectURIs() []string {
	return []string{}
}

func (c *ConfClient) LoginURL(id string) string {
	// return "authorize/callback?id=" + id
	return "login?id=" + id
}

func (c *ConfClient) ApplicationType() op.ApplicationType {
	return c.applicationType
}

func (c *ConfClient) AuthMethod() oidc.AuthMethod {
	return c.authMethod
}

func (c *ConfClient) IDTokenLifetime() time.Duration {
	return 60 * time.Minute
}

func (c *ConfClient) AccessTokenType() op.AccessTokenType {
	return c.accessTokenType
}

func (c *ConfClient) ResponseTypes() []oidc.ResponseType {
	return c.responseTypes
}

func (c *ConfClient) GrantTypes() []oidc.GrantType {
	return c.grantTypes
}

func (c *ConfClient) DevMode() bool {
	return c.devMode
}

func (c *ConfClient) AllowedScopes() []string {
	return nil
}

func (c *ConfClient) RestrictAdditionalIdTokenScopes() func(scopes []string) []string {
	return func(scopes []string) []string {
		return scopes
	}
}

func (c *ConfClient) RestrictAdditionalAccessTokenScopes() func(scopes []string) []string {
	return func(scopes []string) []string {
		return scopes
	}
}

func (c *ConfClient) IsScopeAllowed(scope string) bool {
	return false
}

func (c *ConfClient) IDTokenUserinfoClaimsAssertion() bool {
	return false
}

func (c *ConfClient) ClockSkew() time.Duration {
	return 0
}
