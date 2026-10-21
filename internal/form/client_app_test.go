package form

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/pkg/authn"
)

// newAppClientFlags returns a flag set with the flags read by AddClientFromCli and ModClientFromCli.
func newAppClientFlags(args ...string) *cli.Context {
	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	flags.String("name", "", "Usage")
	flags.String("scope", "", "Usage")
	flags.String("method", "", "Usage")
	flags.Var(cli.NewStringSlice(), "callback", "Usage")
	_ = flags.Parse(args)

	return cli.NewContext(cli.NewApp(), flags, nil)
}

func TestAddClientFromCli_Callback(t *testing.T) {
	t.Run("AppClient", func(t *testing.T) {
		f := AddClientFromCli(newAppClientFlags("--name=Claude", "--callback=https://claude.ai/api/mcp/auth_callback", "--callback= http://127.0.0.1/callback "))
		assert.True(t, f.IsApp())
		assert.Equal(t, authn.MethodPKCE, f.Method())
		assert.Equal(t, "*", f.Scope())
		assert.Equal(t, []string{"https://claude.ai/api/mcp/auth_callback", "http://127.0.0.1/callback"}, f.RedirectURIs)
	})
	t.Run("CustomScope", func(t *testing.T) {
		f := AddClientFromCli(newAppClientFlags("--name=Claude", "--scope=mcp photos", "--callback=https://claude.ai/cb"))
		assert.Equal(t, "mcp photos", f.Scope())
	})
	t.Run("NoCallback", func(t *testing.T) {
		f := AddClientFromCli(newAppClientFlags("--name=Client"))
		assert.False(t, f.IsApp())
		assert.Nil(t, f.RedirectURIs)
		assert.Equal(t, "*", f.Scope())
	})
}

func TestModClientFromCli_Callback(t *testing.T) {
	t.Run("Set", func(t *testing.T) {
		f := ModClientFromCli(newAppClientFlags("--callback=https://claude.ai/cb"))
		assert.Equal(t, []string{"https://claude.ai/cb"}, f.RedirectURIs)
	})
	t.Run("Unset", func(t *testing.T) {
		f := ModClientFromCli(newAppClientFlags("--name=Other"))
		assert.Nil(t, f.RedirectURIs)
	})
}

func TestClient_IsApp(t *testing.T) {
	t.Run("App", func(t *testing.T) {
		assert.True(t, (&Client{AuthMethod: "pkce"}).IsApp())
	})
	t.Run("Client", func(t *testing.T) {
		assert.False(t, (&Client{AuthMethod: "oauth2"}).IsApp())
		assert.False(t, (&Client{}).IsApp())
	})
}

func TestCallbackURLs(t *testing.T) {
	t.Run("Trimmed", func(t *testing.T) {
		assert.Equal(t, []string{"a.b:/c"}, callbackURLs(newAppClientFlags("--callback= a.b:/c ", "--callback=")))
	})
	t.Run("None", func(t *testing.T) {
		assert.Nil(t, callbackURLs(newAppClientFlags()))
	})
}
