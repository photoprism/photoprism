package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestUsersModCommand(t *testing.T) {
	t.Run("ModNotExistingUser", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(UsersModCommand, []string{"mod", "--name=New", "--email=new@test.de", "uqxqg7i1kperxxx0"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)
	})
	t.Run("ModDeletedUser", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(UsersModCommand, []string{"mod", "--name=New", "--email=new@test.de", "deleted"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)
	})
	t.Run("RejectFlagsAfterPositional", func(t *testing.T) {
		// Run with the broken arg order QA reported (positional first, then flags).
		// The stdlib flag parser stops at "alice", so --name / --role would
		// silently no-op without RejectTrailingFlags.
		output, err := RunWithTestContext(UsersModCommand, []string{"mod", "alice", "--name", "Alicia", "--role", "guest"})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "must appear before positional arguments")
		assert.Empty(t, output)

		// Confirm the alice fixture is untouched when it still exists. Earlier
		// tests in the suite may have deleted it, so skip the comparison in
		// that case rather than coupling this test to suite ordering.
		if alice := entity.FindUserByName("alice"); alice != nil {
			assert.Equal(t, "Alice", alice.DisplayName)
			assert.Equal(t, "admin", alice.UserRole)
		}
	})
}
