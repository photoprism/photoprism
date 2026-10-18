package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestAuthResetCommand(t *testing.T) {
	t.Run("NotConfirmed", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_CLI", "")

		// Run command with test context.
		output0, err := RunWithTestContext(AuthListCommand, []string{"ls"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output0, "alice")
		assert.Contains(t, output0, "visitor")

		// Without a terminal the prompt cannot run, which is a usage error rather than a refusal.
		_, err = RunWithTestContext(AuthResetCommand, []string{"reset"})
		assertExitCode(t, err, 2)
		assert.Contains(t, err.Error(), "--yes")

		// Run command with test context.
		output1, err := RunWithTestContext(AuthListCommand, []string{"ls"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output1, "alice")
		assert.Contains(t, output1, "visitor")
	})
	t.Run("Confirmed", func(t *testing.T) {
		c := resetConfigAndOpenDB(t)
		seedAppGrant(t)

		_, err := RunWithTestContext(AuthResetCommand, []string{"reset", "--yes"})
		require.NoError(t, err)

		c = reopenConnection()

		var sessions int64
		require.NoError(t, c.Db().Model(&entity.Session{}).Count(&sessions).Error)
		assert.Zero(t, sessions)

		grants, codes := countAppGrants(t, c.Db())
		assert.Zero(t, grants)
		assert.Zero(t, codes)
	})
}
