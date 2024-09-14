package commands

import (
	"testing"

	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/stretchr/testify/assert"
)

func TestClientsResetCommand(t *testing.T) {
	t.Run("NotConfirmed", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		ctx0 := NewTestContext([]string{"ls"})

		// Run command with test context.
		output0 := capture.Output(func() {
			err = ClientsListCommand.Run(ctx0)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output0, "alice")
		assert.Contains(t, output0, "metrics")

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"reset"})

		// Run command with test context.
		output := capture.Output(func() {
			err = ClientsResetCommand.Run(ctx)
		})

		// Check command output for plausibility.
		//t.Logf(output)
		assert.NoError(t, err)
		assert.Empty(t, output)

		// Run command with test context.
		output1 := capture.Output(func() {
			err = ClientsListCommand.Run(ctx0)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output1, "alice")
		assert.Contains(t, output1, "metrics")
	})
}
