package commands

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestClientsResetCommand(t *testing.T) {
	t.Run("NotConfirmed", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_CLI", "")

		// Run command with test context.
		output0, err := RunWithTestContext(ClientsListCommand, []string{"ls"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output0, "alice")
		assert.Contains(t, output0, "metrics")

		// Without a terminal the prompt cannot run, which is a usage error rather than a refusal.
		_, err = RunWithTestContext(ClientsResetCommand, []string{"reset"})
		assertExitCode(t, err, 2)
		assert.Contains(t, err.Error(), "--yes")

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

		clientSessID := entity.SessionFixtures.Get("client_metrics").ID
		appPasswordID := entity.SessionFixtures.Get("alice_app_password").ID

		// A session naming a client that no longer exists is deleted as well.
		orphan := entity.NewSession(3600, 0)
		orphan.ClientUID = rnd.GenerateUID(entity.ClientUID)
		orphan.ClientName = "orphan-client"

		if err = orphan.Save(); err != nil {
			t.Fatal(err)
		}

		orphanID := orphan.ID

		t.Cleanup(func() { entity.UnscopedDb().Delete(&entity.Session{}, "id = ?", orphanID) })

		assert.Equal(t, 1, countSessionRows(t, clientSessID))
		assert.Equal(t, 1, countSessionRows(t, appPasswordID))
		assert.Equal(t, 1, countSessionRows(t, orphanID))

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

		// Put the client and session fixtures back for the tests that follow.
		c := reopenConnection()
		entity.SetDbProvider(c)

		// The access tokens issued to clients are deleted with them, while app passwords, which
		// belong to user accounts, are kept. Counted in the table, as FindSession may answer from
		// the session cache.
		assert.Equal(t, 0, countSessionRows(t, clientSessID))
		assert.Equal(t, 1, countSessionRows(t, appPasswordID))
		assert.Equal(t, 0, countSessionRows(t, orphanID))

		entity.CreateClientFixtures()
		entity.CreateSessionFixtures()

		// Run command with test context.
		output2, err := RunWithTestContext(ClientsListCommand, []string{"ls"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output2, "alice")
		assert.Contains(t, output2, "metrics")

	})
}

// countSessionRows returns the number of stored sessions with the specified ID.
func countSessionRows(t *testing.T, id string) (n int) {
	t.Helper()

	if err := entity.Db().Model(&entity.Session{}).Where("id = ?", id).Count(&n).Error; err != nil {
		t.Fatal(err)
	}

	return n
}
