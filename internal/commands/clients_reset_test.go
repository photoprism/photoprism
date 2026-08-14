package commands

import (
	"os"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestClientsResetCommand(t *testing.T) {
	t.Run("NotConfirmed", func(t *testing.T) {
		// Run command with test context.
		output0, err := RunWithTestContext(ClientsListCommand, []string{"ls"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output0, "alice")
		assert.Contains(t, output0, "metrics")

		// Run command with test context.
		output, err := RunWithTestContext(ClientsResetCommand, []string{"reset"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Empty(t, output)

		// Run command with test context.
		output1, err := RunWithTestContext(ClientsListCommand, []string{"ls"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output1, "alice")
		assert.Contains(t, output1, "metrics")
	})
	t.Run("Confirmed", func(t *testing.T) {
		_ = os.Setenv(config.EnvVar("cli"), "noninteractive")
		defer os.Unsetenv(config.EnvVar("cli"))
		// Run command with test context.
		output0, err := RunWithTestContext(ClientsListCommand, []string{"ls"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output0, "alice")
		assert.Contains(t, output0, "metrics")

		// Run command with test context.
		output, err := RunWithTestContext(ClientsResetCommand, []string{"reset"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Empty(t, output)

		// Run command with test context.
		output1, err := RunWithTestContext(ClientsListCommand, []string{"ls"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.NotContains(t, output1, "alice")
		assert.NotContains(t, output1, "metrics")

		// Put the clients back
		c := reopenConnection()
		entity.SetDbProvider(c)
		entity.CreateClientFixtures()

		// Run command with test context.
		output2, err := RunWithTestContext(ClientsListCommand, []string{"ls"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output2, "alice")
		assert.Contains(t, output2, "metrics")

	})
}
