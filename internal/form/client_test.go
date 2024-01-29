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

	t.Run("Success", func(t *testing.T) {
		// Create new context with flags.
		ctx := cli.NewContext(cli.NewApp(), flags, nil)

		// Set flag values.
		assert.NoError(t, ctx.Set("name", "Test"))
		assert.NoError(t, ctx.Set("scope", "*"))
		assert.NoError(t, ctx.Set("provider", "client_credentials"))
		assert.NoError(t, ctx.Set("method", "totp"))

		t.Logf("ARGS: %#v", ctx.Args())

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
}

func TestModClientFromCli(t *testing.T) {
	// Specify command flags.
	flags := flag.NewFlagSet("test", 0)
	flags.String("name", "(default)", "Usage")
	flags.String("scope", "(default)", "Usage")
	flags.String("provider", "(default)", "Usage")
	flags.String("method", "(default)", "Usage")

	t.Run("Success", func(t *testing.T) {
		// Create new context with flags.
		ctx := cli.NewContext(cli.NewApp(), flags, nil)

		// Set flag values.
		assert.NoError(t, ctx.Set("name", "Test"))
		assert.NoError(t, ctx.Set("scope", "*"))
		assert.NoError(t, ctx.Set("provider", "client_credentials"))
		assert.NoError(t, ctx.Set("method", "totp"))

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
	})
}
