package form

import (
	"flag"
	"testing"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/pkg/authn"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	t.Run("Defaults", func(t *testing.T) {
		client := NewClient()
		assert.Equal(t, authn.MethodOAuth2, client.Method())
		assert.Equal(t, "", client.Scope())
		assert.Equal(t, "", client.Name())
	})
}

func TestNewClientFromCli(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		globalSet := flag.NewFlagSet("test", 0)
		globalSet.String("name", "Test", "")
		globalSet.String("scope", "*", "")
		globalSet.String("method", "totp", "")

		app := cli.NewApp()
		app.Version = "0.0.0"

		c := cli.NewContext(app, globalSet, nil)

		client := NewClientFromCli(c)
		assert.Equal(t, authn.Method2FA, client.Method())
		assert.Equal(t, "webdav", client.Scope())
		assert.Equal(t, "Test", client.Name())
	})
}
