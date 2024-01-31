package form

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/pkg/authn"
)

func TestNewClient(t *testing.T) {
	t.Run("Defaults", func(t *testing.T) {
		client := NewClient()
		assert.Equal(t, authn.ProviderClientCredentials, client.Provider())
		assert.Equal(t, authn.MethodOAuth2, client.Method())
		assert.Equal(t, "", client.Scope())
		assert.Equal(t, "", client.Name())
	})
}

func TestAddClientFromCli(t *testing.T) {
	// Specify command flags.
	flags := flag.NewFlagSet("test", 0)
	flags.String("name", "(default)", "Usage")
	flags.String("scope", "(default)", "Usage")
	flags.String("provider", "(default)", "Usage")
	flags.String("method", "(default)", "Usage")
	flags.String("id", "(default)", "Usage")
	flags.String("secret", "(default)", "Usage")

	t.Run("Success", func(t *testing.T) {
		// Create new context with flags.
		ctx := cli.NewContext(cli.NewApp(), flags, nil)

		// Set flag values.
		assert.NoError(t, ctx.Set("name", "Test"))
		assert.NoError(t, ctx.Set("scope", "*"))
		assert.NoError(t, ctx.Set("provider", "client_credentials"))
		assert.NoError(t, ctx.Set("method", "totp"))

		//t.Logf("ARGS: %#v", ctx.Args())

		// Check flag values.
		assert.True(t, ctx.IsSet("name"))
		assert.Equal(t, "Test", ctx.String("name"))
		assert.True(t, ctx.IsSet("provider"))
		assert.Equal(t, "client_credentials", ctx.String("provider"))

		// Set form values.
		client := AddClientFromCli(ctx)

		// Check form values.
		assert.Equal(t, authn.ProviderClientCredentials, client.Provider())
		assert.Equal(t, authn.MethodTOTP, client.Method())
		assert.Equal(t, "*", client.Scope())
		assert.Equal(t, "Test", client.Name())
	})
	t.Run("SetID", func(t *testing.T) {
		// Create new context with flags.
		ctx := cli.NewContext(cli.NewApp(), flags, nil)

		// Set flag values.
		assert.NoError(t, ctx.Set("id", "cs5cpu17n6gj8t5r"))
		assert.NoError(t, ctx.Set("name", ""))
		assert.NoError(t, ctx.Set("secret", "xcCbOrw6I0vcoXzhnOmXhjpVSyFq9i8u"))
		assert.NoError(t, ctx.Set("scope", ""))
		assert.NoError(t, ctx.Set("provider", ""))
		assert.NoError(t, ctx.Set("method", ""))

		//t.Logf("ARGS: %#v", ctx.Args())

		// Check flag values.
		assert.True(t, ctx.IsSet("id"))
		assert.True(t, ctx.IsSet("secret"))
		assert.True(t, ctx.IsSet("provider"))

		// Set form values.
		client := AddClientFromCli(ctx)

		// Check form values.
		assert.Equal(t, authn.ProviderDefault, client.Provider())
		assert.Equal(t, authn.MethodDefault, client.Method())
		assert.Equal(t, "*", client.Scope())
		assert.NotEmpty(t, client.Name())
		assert.Equal(t, "xcCbOrw6I0vcoXzhnOmXhjpVSyFq9i8u", client.Secret())
		assert.Equal(t, "cs5cpu17n6gj8t5r", client.ID())
	})
}

func TestModClientFromCli(t *testing.T) {
	// Specify command flags.
	flags := flag.NewFlagSet("test", 0)
	flags.String("name", "(default)", "Usage")
	flags.String("scope", "(default)", "Usage")
	flags.String("provider", "(default)", "Usage")
	flags.String("method", "(default)", "Usage")
	flags.String("role", "(default)", "Usage")
	flags.String("secret", "(default)", "Usage")
	flags.String("expires", "(default)", "Usage")
	flags.String("tokens", "(default)", "Usage")

	t.Run("Success", func(t *testing.T) {
		// Create new context with flags.
		ctx := cli.NewContext(cli.NewApp(), flags, nil)

		// Set flag values.
		assert.NoError(t, ctx.Set("name", "Test"))
		assert.NoError(t, ctx.Set("scope", "*"))
		assert.NoError(t, ctx.Set("provider", "client_credentials"))
		assert.NoError(t, ctx.Set("method", "totp"))
		assert.NoError(t, ctx.Set("role", "visitor"))
		assert.NoError(t, ctx.Set("secret", "xcCbOrw6I0vcoXzhnOmXhjpVSyFq9ijh"))
		assert.NoError(t, ctx.Set("expires", "600"))
		assert.NoError(t, ctx.Set("tokens", "2"))

		// Check flag values.
		assert.True(t, ctx.IsSet("name"))
		assert.Equal(t, "Test", ctx.String("name"))
		assert.True(t, ctx.IsSet("provider"))
		assert.Equal(t, "client_credentials", ctx.String("provider"))

		// Set form values.
		client := ModClientFromCli(ctx)

		// Check form values.
		assert.Equal(t, authn.ProviderClientCredentials, client.Provider())
		assert.Equal(t, authn.MethodTOTP, client.Method())
		assert.Equal(t, "*", client.Scope())
		assert.Equal(t, "Test", client.Name())
		assert.Equal(t, "visitor", client.Role())
		assert.Equal(t, "xcCbOrw6I0vcoXzhnOmXhjpVSyFq9ijh", client.Secret())
		assert.Equal(t, int64(600), client.Expires())
		assert.Equal(t, int64(2), client.Tokens())
	})
}

func TestClient_Expires(t *testing.T) {
	t.Run("ToSmall", func(t *testing.T) {
		c := Client{AuthExpires: -1}
		assert.Equal(t, int64(3600), c.Expires())
	})
	t.Run("ToBig", func(t *testing.T) {
		c := Client{AuthExpires: 999999999}
		assert.Equal(t, int64(2678400), c.Expires())
	})
}

func TestClient_Tokens(t *testing.T) {
	t.Run("ToSmall", func(t *testing.T) {
		c := Client{AuthTokens: -3}
		assert.Equal(t, int64(-1), c.Tokens())
	})
	t.Run("ToBig", func(t *testing.T) {
		c := Client{AuthTokens: 9147483647}
		assert.Equal(t, int64(2147483647), c.Tokens())
	})
}
