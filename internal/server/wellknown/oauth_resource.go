package wellknown

import (
	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
)

// OAuthProtectedResource represents the protected resource metadata of the MCP endpoint (RFC 9728).
type OAuthProtectedResource struct {
	Resource               string   `json:"resource"`
	AuthorizationServers   []string `json:"authorization_servers"`
	ScopesSupported        []string `json:"scopes_supported"`
	BearerMethodsSupported []string `json:"bearer_methods_supported"`
	ResourceName           string   `json:"resource_name,omitempty"`
}

// NewOAuthProtectedResource creates the protected resource metadata of the MCP endpoint based on the config provided.
func NewOAuthProtectedResource(conf *config.Config) *OAuthProtectedResource {
	return &OAuthProtectedResource{
		Resource:               conf.OAuthResourceUrl(),
		AuthorizationServers:   []string{conf.SiteUrl()},
		ScopesSupported:        []string{acl.ResourceMCP.String()},
		BearerMethodsSupported: []string{"header"},
		ResourceName:           conf.SiteTitle(),
	}
}
