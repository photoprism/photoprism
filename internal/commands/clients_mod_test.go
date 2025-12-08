package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientsModCommand(t *testing.T) {
	t.Run("ModNotExistingClient", func(t *testing.T) {
		output, err := RunWithTestContext(ClientsModCommand, []string{"mod", "--name=New", "--scope=test", "cs5cpu17n6gjxxxx"})

		// Check command output for plausibility.
		assert.Error(t, err)
		assert.Empty(t, output)
	})
	t.Run("DisableEnableAuth", func(t *testing.T) {
		// Run command with test context.
		output0, err := RunWithTestContext(ClientsShowCommand, []string{"show", "cs7pvt5h8rw9aaqj"})

		// Check command output for plausibility.
		// t.Logf(output0)
		assert.NoError(t, err)
		assert.Contains(t, output0, "AuthEnabled  │ true")
		assert.Contains(t, output0, "oauth2")

		// Run command with test context.
		output, err := RunWithTestContext(ClientsModCommand, []string{"mod", "--disable", "cs7pvt5h8rw9aaqj"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Empty(t, output)

		// Run command with test context.
		output1, err := RunWithTestContext(ClientsShowCommand, []string{"show", "cs7pvt5h8rw9aaqj"})

		// Check command output for plausibility.
		// t.Logf(output1)
		assert.NoError(t, err)
		assert.Contains(t, output1, "AuthEnabled  │ false")

		// Run command with test context.
		output2, err := RunWithTestContext(ClientsModCommand, []string{"mod", "--enable", "cs7pvt5h8rw9aaqj"})

		// Check command output for plausibility.
		assert.NoError(t, err)
		assert.Empty(t, output2)

		// Run command with test context.
		output3, err := RunWithTestContext(ClientsShowCommand, []string{"show", "cs7pvt5h8rw9aaqj"})

		// Check command output for plausibility.
		// t.Logf(output3)
		assert.NoError(t, err)
		assert.Contains(t, output3, "│ AuthEnabled  │ true ")
	})
	t.Run("RegenerateSecret", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(ClientsModCommand, []string{"mod", "--regenerate", "cs7pvt5h8rw9aaqj"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "Client Secret")
	})
}

func TestClientsModCommand_ModRoleScopeLimits(t *testing.T) {
	// Modify existing fixture client "analytics" (cs7pvt5h8rw9aaqj).
	out0, err := RunWithTestContext(ClientsShowCommand, []string{"show", "cs7pvt5h8rw9aaqj"})
	assert.NoError(t, err)
	assert.Contains(t, out0, "ClientRole")

	// Apply changes.
	_, err = RunWithTestContext(ClientsModCommand, []string{"mod", "--role=portal", "--scope=audit metrics", "--expires=600", "--tokens=3", "cs7pvt5h8rw9aaqj"})
	assert.NoError(t, err)

	// Verify via show.
	out1, err := RunWithTestContext(ClientsShowCommand, []string{"show", "cs7pvt5h8rw9aaqj"})
	assert.NoError(t, err)
	assert.Contains(t, out1, "ClientRole   │ \"portal\"")
	assert.Contains(t, out1, "AuthScope    │ \"audit metrics\"")
	assert.Contains(t, out1, "AuthExpires  │ 600")
	assert.Contains(t, out1, "AuthTokens   │ 3")
}

func TestClientsModCommand_ModRoleToNoneAndEmpty(t *testing.T) {
	// Set to explicit none
	_, err := RunWithTestContext(ClientsModCommand, []string{"mod", "--role=none", "cs7pvt5h8rw9aaqj"})
	assert.NoError(t, err)
	out1, err := RunWithTestContext(ClientsShowCommand, []string{"show", "cs7pvt5h8rw9aaqj"})
	assert.NoError(t, err)
	// Expect empty string value for ClientRole in report output
	assert.Contains(t, out1, "ClientRole   │ \"\"")

	// Set to explicit empty string (treated as none)
	_, err = RunWithTestContext(ClientsModCommand, []string{"mod", "--role=", "cs7pvt5h8rw9aaqj"})
	assert.NoError(t, err)
	out2, err := RunWithTestContext(ClientsShowCommand, []string{"show", "cs7pvt5h8rw9aaqj"})
	assert.NoError(t, err)
	assert.Contains(t, out2, "ClientRole   │ \"\"")

	// Restore to client for other tests
	_, err = RunWithTestContext(ClientsModCommand, []string{"mod", "--role=client", "cs7pvt5h8rw9aaqj"})
	assert.NoError(t, err)
}
