package wellknown

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/config"
)

var (
	OAuthResponseTypes                 = []string{"token"}
	OAuthGrantTypes                    = []string{"client_credentials"}
	OAuthTokenEndpointAuthMethods      = []string{"client_secret_basic", "client_secret_post"}
	OAuthRevocationEndpointAuthMethods = []string{"none"}
)

// OAuthAuthorizationServer represents the values returned by the "/.well-known/oauth-authorization-server" endpoint.
type OAuthAuthorizationServer struct {
	Issuer                                    string   `json:"issuer"`
	AuthorizationEndpoint                     string   `json:"authorization_endpoint"`
	TokenEndpoint                             string   `json:"token_endpoint"`
	RegistrationEndpoint                      string   `json:"registration_endpoint"`
	ResponseTypesSupported                    []string `json:"response_types_supported"`
	ResponseModesSupported                    []string `json:"response_modes_supported"`
	GrantTypesSupported                       []string `json:"grant_types_supported"`
	SubjectTypesSupported                     []string `json:"subject_types_supported"`
	ScopesSupported                           []string `json:"scopes_supported"`
	TokenEndpointAuthMethodsSupported         []string `json:"token_endpoint_auth_methods_supported"`
	ClaimsSupported                           []string `json:"claims_supported"`
	CodeChallengeMethodsSupported             []string `json:"code_challenge_methods_supported"`
	IntrospectionEndpoint                     string   `json:"introspection_endpoint"`
	IntrospectionEndpointAuthMethodsSupported []string `json:"introspection_endpoint_auth_methods_supported"`
	RevocationEndpoint                        string   `json:"revocation_endpoint"`
	RevocationEndpointAuthMethodsSupported    []string `json:"revocation_endpoint_auth_methods_supported"`
	EndSessionEndpoint                        string   `json:"end_session_endpoint"`
	RequestParameterSupported                 bool     `json:"request_parameter_supported"`
	RequestObjectSigningAlgValuesSupported    []string `json:"request_object_signing_alg_values_supported"`
	DeviceAuthorizationEndpoint               string   `json:"device_authorization_endpoint"`
	DpopSigningAlgValuesSupported             []string `json:"dpop_signing_alg_values_supported"`
}

// NewOAuthAuthorizationServer creates a service discovery endpoint response based on the config provided.
func NewOAuthAuthorizationServer(conf *config.Config) *OAuthAuthorizationServer {
	return &OAuthAuthorizationServer{
		Issuer:                                    conf.SiteUrl(),
		AuthorizationEndpoint:                     "",
		TokenEndpoint:                             fmt.Sprintf("%sapi/v1/oauth/token", conf.SiteUrl()),
		ScopesSupported:                           acl.Rules.Resources(),
		ResponseTypesSupported:                    OAuthResponseTypes,
		GrantTypesSupported:                       OAuthGrantTypes,
		TokenEndpointAuthMethodsSupported:         OAuthTokenEndpointAuthMethods,
		ResponseModesSupported:                    []string{},
		SubjectTypesSupported:                     []string{},
		ClaimsSupported:                           []string{},
		CodeChallengeMethodsSupported:             []string{},
		IntrospectionEndpointAuthMethodsSupported: []string{},
		RevocationEndpoint:                        fmt.Sprintf("%sapi/v1/oauth/revoke", conf.SiteUrl()),
		RevocationEndpointAuthMethodsSupported:    OAuthRevocationEndpointAuthMethods,
		EndSessionEndpoint:                        "",
		RequestParameterSupported:                 false,
		RequestObjectSigningAlgValuesSupported:    []string{},
		DeviceAuthorizationEndpoint:               "",
		DpopSigningAlgValuesSupported:             []string{},
	}
}
