package commands

import (
	"testing"

	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/stretchr/testify/assert"
)

func TestClientsAddCommand(t *testing.T) {
	t.Run("AddClient", func(t *testing.T) {
		var err error

		args := []string{"add", "--name=Clara Client", "--scope=photos albums", "--expires=5000", "--tokens=2", "clara"}
		ctx := NewTestContext(args)

		// Run command with test context.
		output := capture.Output(func() {
			err = ClientsAddCommand.Run(ctx, args...)
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
