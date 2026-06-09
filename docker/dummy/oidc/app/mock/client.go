package mock

import (
	"time"

	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/oidc/v3/pkg/op"
)

// defaultRedirectURIs lists the redirect URIs accepted by the dummy provider.
// These match the development origins used by PhotoPrism's local compose stack.
var defaultRedirectURIs = []string{
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

// ConfClient is the dummy provider's op.Client implementation. It is permissive
// by design: any client_id is accepted and treated as a public web client so
// that local development flows can authenticate without prior registration.
type ConfClient struct {
	id              string
	applicationType op.ApplicationType
	authMethod      oidc.AuthMethod
	accessTokenType op.AccessTokenType
	responseTypes   []oidc.ResponseType
	grantTypes      []oidc.GrantType
	devMode         bool
}

// NewClient returns a ConfClient configured for the given client_id.
// The web/native variants mirror the original v1 dummy behavior; everything
// else falls back to a permissive user-agent client.
func NewClient(id string) *ConfClient {
	c := &ConfClient{
		id:      id,
		devMode: true,
		grantTypes: []oidc.GrantType{
			oidc.GrantTypeCode,
			oidc.GrantTypeRefreshToken,
		},
	}
	switch id {
	case "native":
		c.applicationType = op.ApplicationTypeNative
		c.authMethod = oidc.AuthMethodBasic
		c.accessTokenType = op.AccessTokenTypeBearer
		c.responseTypes = []oidc.ResponseType{
			oidc.ResponseTypeCode,
			oidc.ResponseTypeIDToken,
			oidc.ResponseTypeIDTokenOnly,
		}
	case "user_agent":
		c.applicationType = op.ApplicationTypeUserAgent
		c.authMethod = oidc.AuthMethodNone
		c.accessTokenType = op.AccessTokenTypeJWT
		c.responseTypes = []oidc.ResponseType{
			oidc.ResponseTypeIDToken,
			oidc.ResponseTypeIDTokenOnly,
		}
	default:
		// Default to a permissive confidential web client so that any unknown
		// client_id (such as the ones PhotoPrism generates for its backend)
		// can complete the authorization-code flow with client_secret_basic
		// authentication and without PKCE, matching the v1 dummy's behavior.
		c.applicationType = op.ApplicationTypeWeb
		c.authMethod = oidc.AuthMethodBasic
		c.accessTokenType = op.AccessTokenTypeBearer
		c.responseTypes = []oidc.ResponseType{oidc.ResponseTypeCode}
	}
	return c
}

// GetID returns the client_id.
func (c *ConfClient) GetID() string {
	return c.id
}

// RedirectURIs returns the registered redirect URIs.
func (c *ConfClient) RedirectURIs() []string {
	return defaultRedirectURIs
}

// PostLogoutRedirectURIs returns the registered post-logout redirect URIs.
func (c *ConfClient) PostLogoutRedirectURIs() []string {
	return []string{}
}

// LoginURL points the OpenID provider at the dummy's login endpoint.
func (c *ConfClient) LoginURL(id string) string {
	return "/login?id=" + id
}

// ApplicationType returns the OAuth application type.
func (c *ConfClient) ApplicationType() op.ApplicationType {
	return c.applicationType
}

// AuthMethod returns the client's auth method.
func (c *ConfClient) AuthMethod() oidc.AuthMethod {
	return c.authMethod
}

// IDTokenLifetime is the issued id_token lifetime.
func (c *ConfClient) IDTokenLifetime() time.Duration {
	return 60 * time.Minute
}

// AccessTokenType returns the access token format.
func (c *ConfClient) AccessTokenType() op.AccessTokenType {
	return c.accessTokenType
}

// ResponseTypes returns the supported response_type values.
func (c *ConfClient) ResponseTypes() []oidc.ResponseType {
	return c.responseTypes
}

// GrantTypes returns the supported grant_type values.
func (c *ConfClient) GrantTypes() []oidc.GrantType {
	return c.grantTypes
}

// DevMode permits non-compliant configurations so http callbacks are accepted.
func (c *ConfClient) DevMode() bool {
	return c.devMode
}

// AllowedScopes returns nil to indicate that no scope restriction applies.
func (c *ConfClient) AllowedScopes() []string {
	return nil
}

// RestrictAdditionalIdTokenScopes returns the input scopes unchanged.
func (c *ConfClient) RestrictAdditionalIdTokenScopes() func(scopes []string) []string {
	return func(scopes []string) []string { return scopes }
}

// RestrictAdditionalAccessTokenScopes returns the input scopes unchanged.
func (c *ConfClient) RestrictAdditionalAccessTokenScopes() func(scopes []string) []string {
	return func(scopes []string) []string { return scopes }
}

// IsScopeAllowed permits any scope so callers can extend the dummy easily.
func (c *ConfClient) IsScopeAllowed(scope string) bool {
	return true
}

// IDTokenUserinfoClaimsAssertion forces userinfo claims into the id_token to
// match the original v1 dummy's behavior.
func (c *ConfClient) IDTokenUserinfoClaimsAssertion() bool {
	return true
}

// ClockSkew returns no clock skew tolerance.
func (c *ConfClient) ClockSkew() time.Duration {
	return 0
}
