package form

import (
	"flag"
	"testing"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/pkg/authn"

	"github.com/stretchr/testify/assert"
)

func TestUser(t *testing.T) {
	form := &User{UserName: "foobar", UserEmail: "test@test.com", Password: "passwd"}

	assert.Equal(t, "foobar", form.UserName)
	assert.Equal(t, "test@test.com", form.UserEmail)
	assert.Equal(t, "passwd", form.Password)
}

func TestUser_Username(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		form := &User{UserName: "", UserEmail: "test@test.com", Password: "passwd"}
		assert.Equal(t, "", form.Username())
	})
	t.Run("Valid", func(t *testing.T) {
		form := &User{UserName: "foobar", UserEmail: "test@test.com", Password: "passwd"}
		assert.Equal(t, "foobar", form.Username())
	})
	t.Run("Invalid", func(t *testing.T) {
		form := &User{UserName: " Foo Bar4w45 !", UserEmail: "test@test.com", Password: "passwd"}
		assert.Equal(t, "foo bar4w45 !", form.Username())
	})
}

func TestUser_Provider(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		form := &User{UserName: "", UserEmail: "test@test.com", Password: "passwd", AuthProvider: ""}
		assert.Equal(t, authn.ProviderDefault, form.Provider())
	})
	t.Run("Valid", func(t *testing.T) {
		form := &User{UserName: "John", UserEmail: "test@test.com", Password: "passwd", AuthProvider: "local"}
		assert.Equal(t, authn.ProviderLocal, form.Provider())
	})
}

func TestUser_Method(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		form := &User{UserName: "", UserEmail: "test@test.com", Password: "passwd", AuthMethod: ""}
		assert.Equal(t, authn.MethodUndefined, form.Method())
	})
	t.Run("Valid", func(t *testing.T) {
		form := &User{UserName: "John", UserEmail: "test@test.com", Password: "passwd", AuthMethod: "2fa"}
		assert.Equal(t, authn.Method2FA, form.Method())
	})
}

func TestUser_Email(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		form := &User{UserName: "foobar", UserEmail: "", Password: "passwd"}
		assert.Equal(t, "", form.Email())
	})
	t.Run("Valid", func(t *testing.T) {
		form := &User{UserName: "foobar", UserEmail: "test@test.com", Password: "passwd"}
		assert.Equal(t, "test@test.com", form.Email())
	})
	t.Run("Invalid", func(t *testing.T) {
		form := &User{UserName: " Foo Bar4w45 !", UserEmail: "  testtest.srg awrygcom  ", Password: "passwd"}
		assert.Equal(t, "", form.Email())
	})
}

func TestUser_Role(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		form := &User{UserName: "", UserEmail: "test@test.com", Password: "passwd", UserRole: ""}
		assert.Equal(t, "", form.Role())
	})
	t.Run("Valid", func(t *testing.T) {
		form := &User{UserName: "John", UserEmail: "test@test.com", Password: "passwd", UserRole: "admin"}
		assert.Equal(t, "admin", form.Role())
	})
	t.Run("Invalid", func(t *testing.T) {
		form := &User{UserName: "John", UserEmail: "test@test.com", Password: "passwd", UserRole: "ad&*min"}
		assert.Equal(t, "admin", form.Role())
	})
}

func TestUser_Attr(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		form := &User{UserName: "", UserEmail: "test@test.com", Password: "passwd", UserAttr: ""}
		assert.Equal(t, "", form.Attr())
	})
}

func TestUser_Scope(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		form := &User{UserScope: ""}
		assert.Equal(t, "", form.Scope())
	})
	t.Run("Sanitized", func(t *testing.T) {
		form := &User{UserScope: " Photos:VIEW   "}
		assert.Equal(t, "photos:view", form.Scope())
	})
}

func TestNewUserFromCli(t *testing.T) {
	newContext := func(t *testing.T, args ...string) *cli.Context {
		flags := flag.NewFlagSet("test", flag.ContinueOnError)
		flags.String("email", "", "Usage")
		flags.String("role", "", "Usage")
		if err := flags.Parse(args); err != nil {
			t.Fatal(err)
		}
		return cli.NewContext(cli.NewApp(), flags, nil)
	}

	t.Run("TrimsUsername", func(t *testing.T) {
		frm := NewUserFromCli(newContext(t, "--email=Jane@Example.com", "--role=admin", "  Jane.Doe  "))
		assert.Equal(t, "Jane.Doe", frm.UserName)
		assert.Equal(t, "jane.doe", frm.Username())
		assert.Equal(t, "jane@example.com", frm.UserEmail)
		assert.Equal(t, "admin", frm.UserRole)
	})
	t.Run("KeepsInput", func(t *testing.T) {
		frm := NewUserFromCli(newContext(t, "\u212Aate"))
		assert.Equal(t, "\u212Aate", frm.UserName)
		assert.Equal(t, "kate", frm.Username())
	})
	t.Run("Empty", func(t *testing.T) {
		frm := NewUserFromCli(newContext(t))
		assert.Equal(t, "", frm.UserName)
	})
}

func TestUser_MCPSetting(t *testing.T) {
	for in, out := range map[int]int{-9: -1, -1: -1, 0: 0, 1: 1, 2: 1} {
		f := &User{MCP: in}
		assert.Equal(t, out, f.MCPSetting(), in)
	}
}

func TestNewUserFromCli_MCP(t *testing.T) {
	newContext := func(t *testing.T, args ...string) *cli.Context {
		set := flag.NewFlagSet("users add", flag.ContinueOnError)
		_ = set.Bool("mcp", false, "")
		if err := set.Parse(args); err != nil {
			t.Fatal(err)
		}
		return cli.NewContext(cli.NewApp(), set, nil)
	}

	t.Run("NotSet", func(t *testing.T) {
		assert.Equal(t, 0, NewUserFromCli(newContext(t, "alice")).MCP)
	})
	t.Run("True", func(t *testing.T) {
		assert.Equal(t, 1, NewUserFromCli(newContext(t, "--mcp", "alice")).MCP)
	})
	t.Run("False", func(t *testing.T) {
		assert.Equal(t, -1, NewUserFromCli(newContext(t, "--mcp=false", "alice")).MCP)
	})
}

func TestCliSetting(t *testing.T) {
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	_ = set.Bool("on", false, "")
	_ = set.Bool("off", true, "")
	_ = set.Bool("unset", true, "")
	if err := set.Parse([]string{"--on", "--off=false"}); err != nil {
		t.Fatal(err)
	}
	ctx := cli.NewContext(cli.NewApp(), set, nil)

	assert.Equal(t, 1, cliSetting(ctx, "on"))
	assert.Equal(t, -1, cliSetting(ctx, "off"))
	assert.Equal(t, 0, cliSetting(ctx, "unset"))
	assert.Equal(t, 0, cliSetting(ctx, "missing"))
}
