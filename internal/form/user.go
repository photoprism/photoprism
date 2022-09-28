package form

import (
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/pkg/clean"
)

// User represents a user account management form.
type User struct {
	UserName    string `json:"Name" yaml:"Name,omitempty"`
	UserEmail   string `json:"Email,omitempty" yaml:"Email,omitempty"`
	DisplayName string `json:"DisplayName,omitempty" yaml:"DisplayName,omitempty"`
	UserRole    string `json:"Role,omitempty" yaml:"Role,omitempty"`
	UserAttr    string `json:"Attr,omitempty" yaml:"Attr,omitempty"`
	SuperAdmin  bool   `json:"SuperAdmin,omitempty" yaml:"SuperAdmin,omitempty"`
	CanLogin    bool   `json:"CanLogin,omitempty" yaml:"CanLogin,omitempty"`
	CanSync     bool   `json:"CanSync,omitempty" yaml:"CanSync,omitempty"`
	Password    string `json:"Password,omitempty" yaml:"Password,omitempty"`
}

// NewUserFromCli creates a new form with values from a CLI context.
func NewUserFromCli(ctx *cli.Context) User {
	return User{
		UserName:    clean.Username(ctx.Args().First()),
		UserEmail:   clean.Email(ctx.String("email")),
		DisplayName: clean.Name(ctx.String("displayname")),
		UserRole:    clean.Role(ctx.String("role")),
		UserAttr:    clean.Attr(ctx.String("attr")),
		SuperAdmin:  ctx.Bool("superadmin"),
		CanLogin:    !ctx.Bool("disable-login"),
		CanSync:     ctx.Bool("can-sync"),
		Password:    clean.Password(ctx.String("password")),
	}
}

// Name returns the sanitized username in lowercase.
func (f *User) Name() string {
	return clean.Username(f.UserName)
}

// Email returns the sanitized email in lowercase.
func (f *User) Email() string {
	return clean.Email(f.UserEmail)
}

// Role returns the sanitized user role string.
func (f *User) Role() string {
	return clean.Role(f.UserRole)
}

// Attr returns the sanitized user account attributes.
func (f *User) Attr() string {
	return clean.Attr(f.UserAttr)
}
