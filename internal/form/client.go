package form

import (
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/time/unix"
)

// Client represents client application settings.
type Client struct {
	UserUID      string `json:"UserUID,omitempty" yaml:"UserUID,omitempty"`
	UserName     string `gorm:"size:64;index;" json:"UserName" yaml:"UserName,omitempty"`
	ClientID     string `json:"ClientID,omitempty" yaml:"ClientID,omitempty"`
	ClientSecret string `json:"ClientSecret,omitempty" yaml:"ClientSecret,omitempty"`
	ClientName   string `json:"ClientName,omitempty" yaml:"ClientName,omitempty"`
	ClientRole   string `json:"ClientRole,omitempty" yaml:"ClientRole,omitempty"`
	AuthProvider string `json:"AuthProvider,omitempty" yaml:"AuthProvider,omitempty"`
	AuthMethod   string `json:"AuthMethod,omitempty" yaml:"AuthMethod,omitempty"`
	AuthScope    string `json:"AuthScope,omitempty" yaml:"AuthScope,omitempty"`
	AuthExpires  int64  `json:"AuthExpires,omitempty" yaml:"AuthExpires,omitempty"`
	AuthTokens   int64  `json:"AuthTokens,omitempty" yaml:"AuthTokens,omitempty"`
	AuthEnabled  bool   `json:"AuthEnabled,omitempty" yaml:"AuthEnabled,omitempty"`
}

// NewClient creates new client application settings.
func NewClient() Client {
	return Client{
		UserUID:      "",
		UserName:     "",
		ClientID:     "",
		ClientSecret: "",
		ClientName:   "",
		ClientRole:   acl.RoleClient.String(),
		AuthProvider: authn.ProviderClient.String(),
		AuthMethod:   authn.MethodOAuth2.String(),
		AuthScope:    "",
		AuthExpires:  3600,
		AuthTokens:   5,
		AuthEnabled:  true,
	}
}

// AddClientFromCli creates a new form for adding a client with values from the specified CLI context.
func AddClientFromCli(ctx *cli.Context) Client {
	f := NewClient()

	if user := clean.Username(ctx.Args().First()); rnd.IsUID(user, 'u') {
		f.UserUID = user
	} else if user != "" {
		f.UserName = user
	}

	if ctx.IsSet("id") {
		f.ClientID = ctx.String("id")
	}

	if ctx.IsSet("secret") {
		f.ClientSecret = ctx.String("secret")
	}

	if ctx.IsSet("name") {
		f.ClientName = clean.Name(ctx.String("name"))
	}

	if f.ClientName == "" {
		f.ClientName = rnd.Name()
	}

	f.ClientRole = clean.Name(ctx.String("role"))

	if f.ClientRole == "" {
		f.ClientRole = acl.RoleClient.String()
	}

	f.AuthProvider = authn.Provider(ctx.String("provider")).String()

	if f.AuthProvider == "" {
		f.AuthProvider = authn.ProviderClient.String()
	}

	f.AuthMethod = authn.Method(ctx.String("method")).String()

	if f.AuthMethod == "" {
		f.AuthMethod = authn.MethodOAuth2.String()
	}

	f.AuthScope = clean.Scope(ctx.String("scope"))

	if f.AuthScope == "" {
		f.AuthScope = "*"
	}

	f.AuthExpires = ctx.Int64("expires")
	f.AuthTokens = ctx.Int64("tokens")

	return f
}

// ModClientFromCli creates a new form for modifying a client with values from the specified CLI context.
func ModClientFromCli(ctx *cli.Context) Client {
	f := Client{}

	f.ClientID = clean.UID(ctx.Args().First())

	if ctx.IsSet("secret") {
		f.ClientSecret = ctx.String("secret")
	}

	if ctx.IsSet("name") {
		f.ClientName = clean.Name(ctx.String("name"))
	}

	if ctx.IsSet("role") {
		f.ClientRole = clean.Name(ctx.String("role"))
	}

	if ctx.IsSet("provider") {
		f.AuthProvider = authn.Provider(ctx.String("provider")).String()
	}

	if ctx.IsSet("method") {
		f.AuthMethod = authn.Method(ctx.String("method")).String()
	}

	if ctx.IsSet("scope") {
		f.AuthScope = clean.Scope(ctx.String("scope"))
	}

	if ctx.IsSet("expires") {
		f.AuthExpires = ctx.Int64("expires")
	}

	if ctx.IsSet("tokens") {
		f.AuthTokens = ctx.Int64("tokens")
	}

	if ctx.Bool("enable") {
		f.AuthEnabled = true
	} else if ctx.Bool("disable") {
		f.AuthEnabled = false
	}

	return f
}

// ID returns the client id, if any.
func (f *Client) ID() string {
	if !rnd.IsUID(f.ClientID, 'c') {
		return ""
	}

	return f.ClientID
}

// Secret returns the client secret, if any.
func (f *Client) Secret() string {
	if !rnd.IsClientSecret(f.ClientSecret) {
		return ""
	}

	return f.ClientSecret
}

// Name returns the sanitized client name.
func (f *Client) Name() string {
	return clean.Name(f.ClientName)
}

// Role returns the sanitized client role.
func (f *Client) Role() string {
	return clean.Role(f.ClientRole)
}

// Provider returns the sanitized auth provider name.
func (f *Client) Provider() authn.ProviderType {
	return authn.Provider(f.AuthProvider)
}

// Method returns the sanitized auth method name.
func (f *Client) Method() authn.MethodType {
	return authn.Method(f.AuthMethod)
}

// Scope returns the client scopes as sanitized string.
func (f Client) Scope() string {
	return clean.Scope(f.AuthScope)
}

// Expires returns the access token expiry time in seconds or 0 if not specified.
func (f Client) Expires() int64 {
	if f.AuthExpires > unix.Month {
		return unix.Month
	} else if f.AuthExpires > 0 {
		return f.AuthExpires
	} else if f.AuthExpires < 0 {
		return unix.Hour
	}

	return 0
}

// Tokens returns the access token limit or 0 if not specified.
func (f Client) Tokens() int64 {
	if f.AuthTokens > 2147483647 {
		return 2147483647
	} else if f.AuthTokens > 0 {
		return f.AuthTokens
	} else if f.AuthTokens < 0 {
		return -1
	}

	return 0
}
