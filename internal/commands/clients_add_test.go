package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientsAddCommand(t *testing.T) {
	t.Run("AddClient", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(ClientsAddCommand, []string{"add", "--name=Clara Client", "--scope=photos albums", "--expires=5000", "--tokens=2", "clara"})

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

func TestClientsAddCommand_PublicClient(t *testing.T) {
	t.Run("PublicWithRedirectURIs", func(t *testing.T) {
		output, err := RunWithTestContext(ClientsAddCommand, []string{
			"add", "--name=Native App", "--scope=photos", "--public",
			"--redirect-uri=photoprism://callback", "--redirect-uri=https://app.example.com/cb",
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Native App")
		assert.Contains(t, output, "photoprism://callback")
		assert.Contains(t, output, "https://app.example.com/cb")
		// Public clients use PKCE and are not issued a secret.
		assert.Contains(t, output, "public client")
		assert.NotContains(t, output, "Client Secret")
	})
	t.Run("InvalidRedirectURIRejected", func(t *testing.T) {
		_, err := RunWithTestContext(ClientsAddCommand, []string{
			"add", "--name=Bad App", "--scope=photos", "--redirect-uri=not-a-uri",
		})
		assert.Error(t, err)
	})
}
