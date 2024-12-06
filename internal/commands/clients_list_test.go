package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientsListCommand(t *testing.T) {
	t.Run("All", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(ClientsListCommand, []string{"ls"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "alice")
		assert.Contains(t, output, "bob")
		assert.Contains(t, output, "Monitoring")
		assert.NotContains(t, output, "visitor")
	})
	t.Run("Monitoring", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(ClientsListCommand, []string{"ls", "monitoring"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "|  Scope  |")
		assert.Contains(t, output, "Monitoring")
		assert.NotContains(t, output, "alice")
		assert.NotContains(t, output, "bob")
		assert.NotContains(t, output, "visitor")
	})
	t.Run("CSV", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(ClientsListCommand, []string{"ls", "--csv", "monitoring"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "Client ID;Name;Authentication Method;User;Role;Scope;Enabled;Access Token Lifetime;Created At")
		assert.Contains(t, output, "Monitoring")
		assert.Contains(t, output, "metrics")
		assert.NotContains(t, output, "bob")
		assert.NotContains(t, output, "alice")
	})
	t.Run("NoResult", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(ClientsListCommand, []string{"ls", "notexisting"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Empty(t, output)
	})
	t.Run("Error", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(ClientsListCommand, []string{"ls", "--xyz", "monitoring"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Empty(t, output)
		assert.Error(t, err)
	})
}
