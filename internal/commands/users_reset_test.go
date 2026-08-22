package commands

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestUsersResetCommand(t *testing.T) {
	c := resetConfigAndOpenDB()
	// reset as this test removes all users
	defer resetConfigAndDB()

	t.Run("NotConfirmed", func(t *testing.T) {
		// Run command with test context.
		output0, err := RunWithTestContext(UsersListCommand, []string{"ls"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output0, "alice")
		assert.Contains(t, output0, "bob")

		// Run command with test context.
		output, err := RunWithTestContext(UsersResetCommand, []string{"reset"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Empty(t, output)

		// Run command with test context.
		output1, err := RunWithTestContext(UsersListCommand, []string{"ls"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output1, "alice")
		assert.Contains(t, output1, "bob")
	})
	t.Run("Reset", func(t *testing.T) {
		// c := resetConfigAndDB()
		count := int64(0)
		if err := c.Db().Model(&entity.User{}).Count(&count).Error; err != nil {
			assert.NoError(t, err)
			return
		}
		assert.Greater(t, count, int64(3)) // Make sure we have a populated database

		dbDrv := os.Getenv("PHOTOPRISM_TEST_DRIVER")
		dbDSN := os.Getenv("PHOTOPRISM_TEST_DSN")
		// Run command with test context.
		appArgs := []string{"photoprism",
			"--database-driver", dbDrv,
			"--database-dsn", dbDSN}
		cmdArgs := []string{"reset", "--yes"}

		ctx := NewTestContextWithParse(appArgs, cmdArgs)

		// Setup and capture output
		buffer := bytes.Buffer{}
		log.SetOutput(&buffer)

		output, err := RunWithProvidedTestContext(ctx, UsersResetCommand, cmdArgs)
		// Reset logger
		log.SetOutput(os.Stdout)

		// Check command output for plausibility.
		// t.Logf("buffer = %s", buffer.String())
		assert.NoError(t, err)
		assert.Empty(t, output)
		assert.Contains(t, buffer.String(), "the user database has been recreated and is now in a clean state")

		c = reopenConnection()

		count = int64(1)
		if err := c.Db().Model(&entity.User{}).Count(&count).Error; err != nil {
			assert.NoError(t, err)
			return
		}
		assert.Equal(t, int64(0), count)
	})
}
