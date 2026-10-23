package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/authn"
)

func TestClientsAddCommand(t *testing.T) {
	t.Run("AddClient", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(ClientsAddCommand, []string{"add", "--name=Clara Client", "--scope=photos albums", "--expires=5000", "--tokens=2"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "Clara Client")
		assert.Contains(t, output, "client")
		assert.Contains(t, output, "albums photos")
		assert.Contains(t, output, "Client Secret")
	})
}

func TestClientsAddCommand_AddWithRoleAndUser(t *testing.T) {
	t.Run("AddClientWithRolePortalAndUserAlice", func(t *testing.T) {
		output, err := RunWithTestContext(ClientsAddCommand, []string{"add", "--name=Roly Poly", "--scope=vision", "--role=portal", "alice"})

		assert.NoError(t, err)
		assert.Contains(t, output, "Roly Poly")
		assert.Contains(t, output, "portal")
		assert.Contains(t, output, "vision")
		assert.Contains(t, output, "alice")
		assert.Contains(t, output, "Client Secret")
	})
}

func TestClientsAddCommand_App(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		output, err := RunWithTestContext(ClientsAddCommand, []string{"add", "--name=Claude Connector", "--redirect-uri=https://claude.ai/api/mcp/auth_callback", "--redirect-uri=http://127.0.0.1/callback"})

		assert.NoError(t, err)
		assert.Contains(t, output, "Claude Connector")
		assert.Contains(t, output, "PKCE")
		assert.Contains(t, output, "connect the app")
		assert.NotContains(t, output, "Client Secret")
		assert.NotContains(t, output, "tokens")

		found := &entity.Client{}
		require.NoError(t, entity.Db().Where("client_name = ?", "Claude Connector").Order("created_at DESC").First(found).Error)
		assert.True(t, found.IsApp())
		assert.Equal(t, authn.ClientPublic, found.ClientType)
		assert.Equal(t, "*", found.AuthScope)
		assert.Equal(t, []string{"https://claude.ai/api/mcp/auth_callback", "http://127.0.0.1/callback"}, found.RedirectURIs())
		assert.Nil(t, entity.FindPassword(found.ClientUID))
		assert.NoError(t, found.Delete())
	})
	t.Run("InstanceRole", func(t *testing.T) {
		_, err := RunWithTestContext(ClientsAddCommand, []string{"add", "--name=Role App", "--role=instance", "--redirect-uri=https://claude.ai/cb"})
		assert.Error(t, err)
	})
	t.Run("InvalidRedirectURI", func(t *testing.T) {
		_, err := RunWithTestContext(ClientsAddCommand, []string{"add", "--name=Bad App", "--redirect-uri=javascript:alert(1)"})
		assert.Error(t, err)
	})
	t.Run("WithUser", func(t *testing.T) {
		_, err := RunWithTestContext(ClientsAddCommand, []string{"add", "--name=User App", "--redirect-uri=https://claude.ai/cb", "alice"})
		assert.Error(t, err)
	})
}
