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
	t.Run("AdoptOidcIdentityWithIssuer", func(t *testing.T) {
		// Create a fresh local account, then adopt it into an OIDC identity with a
		// pinned issuer — the documented reconcile path for a pre-existing account.
		_, err := RunWithTestContext(UsersAddCommand, []string{"add", "--name=Adopt Me", "--email=adopt@example.com", "--password=test1234", "--role=admin", "adoptme"})
		require.NoError(t, err)

		_, err = RunWithTestContext(UsersModCommand, []string{"mod", "--auth=oidc", "--auth-id=us9k2lqd8m3n7abc", "--auth-issuer=https://portal.example.com/", "adoptme"})
		require.NoError(t, err)

		m := entity.FindUserByName("adoptme")
		require.NotNil(t, m)
		assert.Equal(t, "oidc", m.AuthProvider)
		assert.Equal(t, "us9k2lqd8m3n7abc", m.AuthID)
		assert.Equal(t, "https://portal.example.com/", m.AuthIssuer)
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
