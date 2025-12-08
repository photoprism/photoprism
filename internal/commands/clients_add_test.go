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
