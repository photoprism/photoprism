package wellknown

import (
	"strings"

	"github.com/photoprism/photoprism/internal/config"
)

// Portal OIDC OP metadata, distinct from the generic OAuth2 metadata above.
// These values are dictated by the Portal OIDC spec: ID tokens are signed
// with EdDSA, only the authorization-code flow is supported, PKCE S256 is
// required, and the OIDC standard scope vocabulary applies.
var (
	PortalOIDCResponseTypes            = []string{"code"}
	PortalOIDCGrantTypes               = []string{"authorization_code"}
	PortalOIDCSubjectTypes             = []string{"public"}
	PortalOIDCIDTokenSigningAlgs       = []string{"EdDSA"}
	PortalOIDCScopes                   = []string{"openid", "profile", "email", "cluster", "groups"}
	PortalOIDCCodeChallengeMethods     = []string{"S256"}
	PortalOIDCTokenEndpointAuthMethods = []string{"client_secret_basic", "client_secret_post"}
)

// NewPortalOpenIDConfiguration builds the discovery JSON for the Portal's
// OIDC OP. The returned shape advertises the endpoints, signing algorithm,
// and scope vocabulary documented in specs/portal/cluster-oidc.md; instances
// consume it through their existing OIDC RP code.
func NewPortalOpenIDConfiguration(conf *config.Config) *OpenIDConfiguration {
	issuer := strings.TrimRight(conf.PortalOIDCIssuer(), "/")

	return &OpenIDConfiguration{
		Issuer: issuer + "/",
		// The OP authorize/token/userinfo endpoints live under the APIv1 path, and
		// the JWKS under /.well-known. The bare paths are correct here (not
		// conf.ApiUri() / conf.BaseUri(...)): the issuer already carries any
		// deployment base path, so prefixing with the already-base-pathed accessor
		// would double it for sub-path deployments.
		AuthorizationEndpoint:                     issuer + config.ApiUri + "/oauth/authorize",
		TokenEndpoint:                             issuer + config.ApiUri + "/oauth/token",
		UserinfoEndpoint:                          issuer + config.ApiUri + "/oauth/userinfo",
		JwksUri:                                   issuer + "/.well-known/jwks.json",
		ResponseTypesSupported:                    PortalOIDCResponseTypes,
		GrantTypesSupported:                       PortalOIDCGrantTypes,
		SubjectTypesSupported:                     PortalOIDCSubjectTypes,
		IdTokenSigningAlgValuesSupported:          PortalOIDCIDTokenSigningAlgs,
		ScopesSupported:                           PortalOIDCScopes,
		TokenEndpointAuthMethodsSupported:         PortalOIDCTokenEndpointAuthMethods,
		CodeChallengeMethodsSupported:             PortalOIDCCodeChallengeMethods,
		ResponseModesSupported:                    []string{},
		ClaimsSupported:                           []string{},
		IntrospectionEndpointAuthMethodsSupported: []string{},
		RevocationEndpointAuthMethodsSupported:    []string{},
		RequestObjectSigningAlgValuesSupported:    []string{},
		DpopSigningAlgValuesSupported:             []string{},
	}
}
