package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
)

// TestClientsModCommand checks authentication changes with isolated client and session fixtures.
func TestClientsModCommand(t *testing.T) {
	previous := requireTestDb(t)
	fixture := entity.SessionFixtures.Get("client_analytics")
	t.Cleanup(func() {
		previous.RegisterDb()
		require.NoError(t, previous.Db().Where("id = ?", fixture.ID).First(&entity.Session{}).Error)
	})
	resetConfigAndOpenDB(t)
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

// newDeletedTestClient creates and deletes a client, purges it when the test ends, and returns its UID.
func newDeletedTestClient(t *testing.T, name string) string {
	t.Helper()

	m := entity.NewClient().SetName(name).SetScope("metrics")
	require.NoError(t, m.Create())
	t.Cleanup(func() {
		reopenConnection()
		assert.NoError(t, m.Purge(), "purge test client")
	})
	require.NoError(t, m.Delete())
	require.True(t, entity.FindClient(m.ClientUID).Deleted())

	return m.ClientUID
}

// TestClientsModCommand_RestorePrompt checks when restoring a deleted client asks for confirmation.
func TestClientsModCommand_RestorePrompt(t *testing.T) {
	requireTestDb(t)
	t.Setenv("PHOTOPRISM_CLI", "")

	t.Run("NoTerminal", func(t *testing.T) {
		uid := newDeletedTestClient(t, "RestoreModTty")

		_, err := RunWithTestContext(ClientsModCommand, []string{"mod", "--scope=test", uid})

		require.Error(t, err)
		assert.Equal(t, 1, ExitCode(err))
		assert.Contains(t, err.Error(), "--restore")
		assert.True(t, entity.FindClient(uid).Deleted())
	})
	t.Run("NonInteractiveEnv", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_CLI", NONINTERACTIVE)
		uid := newDeletedTestClient(t, "RestoreModEnv")

		_, err := RunWithTestContext(ClientsModCommand, []string{"mod", "--scope=test", uid})

		require.Error(t, err)
		assert.Equal(t, 1, ExitCode(err))
		assert.Contains(t, err.Error(), "--restore")
		assert.True(t, entity.FindClient(uid).Deleted())
	})
	t.Run("Declined", func(t *testing.T) {
		uid := newDeletedTestClient(t, "RestoreModNo")
		pipeResetAnswers(t, "n\n")

		_, err := RunWithTestContext(ClientsModCommand, []string{"mod", "--scope=test", uid})

		assert.ErrorContains(t, err, "has been deleted")

		m := entity.FindClient(uid)
		assert.True(t, m.Deleted())
		assert.Equal(t, "metrics", m.AuthScope)
	})
	t.Run("Confirmed", func(t *testing.T) {
		uid := newDeletedTestClient(t, "RestoreModYes")
		pipeResetAnswers(t, "y\n")

		_, err := RunWithTestContext(ClientsModCommand, []string{"mod", "--scope=test", uid})

		assert.NoError(t, err)
		assert.False(t, entity.FindClient(uid).Deleted())
	})
	t.Run("RestoreFlag", func(t *testing.T) {
		uid := newDeletedTestClient(t, "RestoreModFlag")

		_, err := RunWithTestContext(ClientsModCommand, []string{"mod", "--restore", uid})

		assert.NoError(t, err)
		assert.False(t, entity.FindClient(uid).Deleted())
	})
	t.Run("PreviousSecretNotRestored", func(t *testing.T) {
		m := entity.NewClient().SetName("RestoreModSecret").SetScope("metrics")
		require.NoError(t, m.Create())
		t.Cleanup(func() {
			reopenConnection()
			assert.NoError(t, m.Purge(), "purge test client")
		})

		secret, err := m.NewSecret()
		require.NoError(t, err)

		// Marks the record as deleted while its secret is still stored.
		require.NoError(t, entity.Db().Delete(m).Error)

		_, err = RunWithTestContext(ClientsModCommand, []string{"mod", "--restore", m.ClientUID})

		require.NoError(t, err)
		assert.False(t, entity.FindClient(m.ClientUID).Deleted())
		assert.False(t, entity.FindClient(m.ClientUID).VerifySecret(secret))
	})
	t.Run("RestoreRegenerate", func(t *testing.T) {
		uid := newDeletedTestClient(t, "RestoreModRegenerate")

		output, err := RunWithTestContext(ClientsModCommand, []string{"mod", "--restore", "--regenerate", uid})

		require.NoError(t, err)
		assert.Contains(t, output, "Client Secret")
		assert.NotNil(t, entity.FindPassword(uid))
	})
}
