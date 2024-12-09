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
