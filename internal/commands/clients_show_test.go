package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientsShowCommand(t *testing.T) {
	t.Run("All", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(ClientsShowCommand, []string{"show", "cs5gfen1bgxz7s9i"})

		// Check command output for plausibility.
		//t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "Alice")
		assert.Contains(t, output, "oauth2")
		assert.Contains(t, output, "confidential")
	})
	t.Run("NoResult", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(ClientsShowCommand, []string{"show", "cs5gfen1bgxzxxxx"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)
	})
}
