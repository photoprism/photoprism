package commands

import (
	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClientsAddCommand(t *testing.T) {
	t.Run("AddClient", func(t *testing.T) {
		var err error

		ctx := NewTestContext([]string{"add", "--name=Clara Client", "--scope=photos albums", "--expires=5000", "--tokens=2", "clara"})

		// Run command with test context.
		output := capture.Output(func() {
			err = ClientsAddCommand.Run(ctx)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "Clara Client")
		assert.Contains(t, output, "client")
		assert.Contains(t, output, "albums photos")
		assert.Contains(t, output, "Client Secret")
	})
}
