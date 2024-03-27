package wellknown

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/config"
)

// OpenIDConfiguration represents the values returned by the "/.well-known/openid-configuration" endpoint.
type OpenIDConfiguration struct {
	Issuer                                    string   `json:"issuer"`
	AuthorizationEndpoint                     string   `json:"authorization_endpoint"`
	TokenEndpoint                             string   `json:"token_endpoint"`
	UserinfoEndpoint                          string   `json:"userinfo_endpoint"`
	RegistrationEndpoint                      string   `json:"registration_endpoint"`
	JwksUri                                   string   `json:"jwks_uri"`
	ResponseTypesSupported                    []string `json:"response_types_supported"`
	ResponseModesSupported                    []string `json:"response_modes_supported"`
	GrantTypesSupported                       []string `json:"grant_types_supported"`
	SubjectTypesSupported                     []string `json:"subject_types_supported"`
	IdTokenSigningAlgValuesSupported          []string `json:"id_token_signing_alg_values_supported"`
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

// NewOpenIDConfiguration creates a service discovery endpoint response based on the config provided.
func NewOpenIDConfiguration(conf *config.Config) *OpenIDConfiguration {
	return &OpenIDConfiguration{
		Issuer:                                    conf.SiteUrl(),
		AuthorizationEndpoint:                     "",
		TokenEndpoint:                             fmt.Sprintf("%sapi/v1/oauth/token", conf.SiteUrl()),
		UserinfoEndpoint:                          "",
		RegistrationEndpoint:                      "",
		JwksUri:                                   "",
		ResponseTypesSupported:                    OAuthResponseTypes,
		ResponseModesSupported:                    []string{},
		GrantTypesSupported:                       OAuthGrantTypes,
		SubjectTypesSupported:                     []string{},
		IdTokenSigningAlgValuesSupported:          []string{},
		ScopesSupported:                           acl.Resources.Resources(),
		TokenEndpointAuthMethodsSupported:         OAuthTokenEndpointAuthMethods,
		ClaimsSupported:                           []string{},
		CodeChallengeMethodsSupported:             []string{},
		IntrospectionEndpoint:                     "",
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
