package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsersLegacyCommand(t *testing.T) {
	t.Run("All", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(UsersLegacyCommand, []string{""})

		// Check command output for plausibility.
		//t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "| ID | UID | Name | User | Email | Admin | Created At |")
	})
}
