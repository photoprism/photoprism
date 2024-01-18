package form

import (
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Client represents client application settings.
type Client struct {
	UserUID     string `json:"UserUID,omitempty" yaml:"UserUID,omitempty"`
	UserName    string `gorm:"size:64;index;" json:"UserName" yaml:"UserName,omitempty"`
	ClientName  string `json:"ClientName,omitempty" yaml:"ClientName,omitempty"`
	ClientRole  string `json:"ClientRole,omitempty" yaml:"ClientRole,omitempty"`
	AuthMethod  string `json:"AuthMethod,omitempty" yaml:"AuthMethod,omitempty"`
	AuthScope   string `json:"AuthScope,omitempty" yaml:"AuthScope,omitempty"`
	AuthExpires int64  `json:"AuthExpires,omitempty" yaml:"AuthExpires,omitempty"`
	AuthTokens  int64  `json:"AuthTokens,omitempty" yaml:"AuthTokens,omitempty"`
	AuthEnabled bool   `json:"AuthEnabled,omitempty" yaml:"AuthEnabled,omitempty"`
}

// NewClient creates new client application settings.
func NewClient() Client {
	return Client{
		UserUID:     "",
		UserName:    "",
		ClientName:  "",
		AuthMethod:  authn.MethodOAuth2.String(),
		AuthScope:   "",
		AuthExpires: 3600,
		AuthTokens:  5,
		AuthEnabled: true,
	}
}

// NewClientFromCli creates a new form with values from a CLI context.
func NewClientFromCli(ctx *cli.Context) Client {
	f := NewClient()

	f.ClientName = clean.Name(ctx.String("name"))
	f.ClientRole = clean.Name(ctx.String("role"))
	f.AuthScope = clean.Scope(ctx.String("scope"))
	f.AuthMethod = authn.Method(ctx.String("method")).String()

	if authn.MethodOAuth2.NotEqual(f.AuthMethod) {
		f.AuthScope = "webdav"
	}

	if user := clean.Username(ctx.Args().First()); rnd.IsUID(user, 'u') {
		f.UserUID = user
	} else if user != "" {
		f.UserName = user
	}

	return f
}

// Name returns the sanitized client name.
func (f *Client) Name() string {
	return clean.Name(f.ClientName)
}

// Role returns the sanitized client role.
func (f *Client) Role() string {
	return clean.Role(f.ClientRole)
}

// Method returns the sanitized auth method name.
func (f *Client) Method() authn.MethodType {
	return authn.Method(f.AuthMethod)
}

// Scope returns the client scopes as sanitized string.
func (f Client) Scope() string {
	return clean.Scope(f.AuthScope)
}
