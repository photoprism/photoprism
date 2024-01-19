package commands

import (
	"testing"

	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/stretchr/testify/assert"
)

func TestClientsShowCommand(t *testing.T) {
	t.Run("All", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"show", "cs5gfen1bgxz7s9i"})

		// Run command with test context.
		output := capture.Output(func() {
			err = ClientsShowCommand.Run(ctx)
		})

		// Check command output for plausibility.
		//t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "Alice")
		assert.Contains(t, output, "oauth2")
		assert.Contains(t, output, "confidential")
	})
	t.Run("NoResult", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"show", "cs5gfen1bgxzxxxx"})

		// Run command with test context.
		output := capture.Output(func() {
			err = ClientsShowCommand.Run(ctx)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)
	})
}
