package form

import (
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
)

// User represents a user account form.
type User struct {
	UserName     string       `json:"Name,omitempty" yaml:"Name,omitempty"`
	AuthProvider string       `json:"AuthProvider,omitempty" yaml:"AuthProvider,omitempty"`
	AuthMethod   string       `json:"AuthMethod,omitempty" yaml:"AuthMethod,omitempty"`
	AuthIssuer   string       `json:"AuthIssuer,omitempty" yaml:"AuthIssuer,omitempty"`
	AuthID       string       `json:"AuthID,omitempty" yaml:"AuthID,omitempty"`
	UserEmail    string       `json:"Email,omitempty" yaml:"Email,omitempty"`
	DisplayName  string       `json:"DisplayName,omitempty" yaml:"DisplayName,omitempty"`
	UserRole     string       `json:"Role,omitempty" yaml:"Role,omitempty"`
	UserScope    string       `json:"Scope,omitempty" yaml:"Scope,omitempty"`
	SuperAdmin   bool         `json:"SuperAdmin,omitempty" yaml:"SuperAdmin,omitempty"`
	CanLogin     bool         `json:"CanLogin,omitempty" yaml:"CanLogin,omitempty"`
	WebDAV       bool         `json:"WebDAV,omitempty" yaml:"WebDAV,omitempty"`
	MCP          int          `json:"MCP,omitempty" yaml:"MCP,omitempty"`
	UserAttr     string       `json:"Attr,omitempty" yaml:"Attr,omitempty"`
	BasePath     string       `json:"BasePath,omitempty" yaml:"BasePath,omitempty"`
	UploadPath   string       `json:"UploadPath,omitempty" yaml:"UploadPath,omitempty"`
	Password     string       `json:"Password,omitempty" yaml:"Password,omitempty"` //nolint:gosec // G117: Expected user credential field.
	DeletedAt    *time.Time   `json:"DeletedAt,omitempty" yaml:"DeletedAt,omitempty"`
	UserDetails  *UserDetails `json:"Details,omitempty"`
}

// NewUserFromCli creates a new form with values from a CLI context. The username is only trimmed, so
// that the rules for new usernames see the input; Username returns its sanitized form.
func NewUserFromCli(ctx *cli.Context) User {
	return User{
		UserName:     strings.TrimSpace(ctx.Args().First()),
		AuthProvider: clean.TypeLower(ctx.String("auth")),
		AuthID:       clean.Auth(ctx.String("auth-id")),
		AuthIssuer:   clean.Uri(ctx.String("auth-issuer")),
		UserEmail:    clean.Email(ctx.String("email")),
		DisplayName:  clean.Name(ctx.String("name")),
		UserRole:     clean.Role(ctx.String("role")),
		UserScope:    clean.Scope(ctx.String("scope")),
		SuperAdmin:   ctx.Bool("superadmin"),
		CanLogin:     !ctx.Bool("no-login"),
		WebDAV:       ctx.Bool("webdav"),
		MCP:          cliSetting(ctx, "mcp"),
		UserAttr:     clean.Attr(ctx.String("attr")),
		BasePath:     clean.UserPath(ctx.String("base-path")),
		UploadPath:   clean.UserPath(ctx.String("upload-path")),
		Password:     clean.Password(ctx.String("password")),
	}
}

// cliSetting returns 1 if the bool flag is set to true, -1 if it is set to false, and 0 if it is not set.
func cliSetting(ctx *cli.Context, name string) int {
	switch {
	case !ctx.IsSet(name):
		return 0
	case ctx.Bool(name):
		return 1
	default:
		return -1
	}
}

// MCPSetting returns the MCP setting clamped to -1 (disabled), 0 (default), or 1 (enabled).
func (f *User) MCPSetting() int {
	switch {
	case f.MCP < 0:
		return -1
	case f.MCP > 0:
		return 1
	default:
		return 0
	}
}

// Username returns the sanitized username in lowercase.
func (f *User) Username() string {
	return clean.Username(f.UserName)
}

// Provider returns the sanitized auth provider name.
func (f *User) Provider() authn.ProviderType {
	return authn.Provider(f.AuthProvider)
}

// Method returns the sanitized auth method name.
func (f *User) Method() authn.MethodType {
	return authn.Method(f.AuthMethod)
}

// Email returns the sanitized email in lowercase.
func (f *User) Email() string {
	return clean.Email(f.UserEmail)
}

// Role returns the sanitized user role string.
func (f *User) Role() string {
	return clean.Role(f.UserRole)
}

// Scope returns the sanitized user scope.
func (f *User) Scope() string {
	return clean.Scope(f.UserScope)
}

// Attr returns the sanitized user account attributes.
func (f *User) Attr() string {
	return clean.Attr(f.UserAttr)
}
