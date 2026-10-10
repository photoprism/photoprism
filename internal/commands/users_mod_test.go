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
		t.Setenv("PHOTOPRISM_CLI", "")

		// Restoring the deleted account needs a confirmation, which cannot be asked without a terminal.
		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--name=New", "--email=new@test.de", "deleted"})

		assert.Equal(t, 1, ExitCode(err))
		assert.True(t, entity.FindUserByName("deleted").IsDeleted())
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
	t.Run("SharedEmail", func(t *testing.T) {
		// Accounts may share an email address, and changing one of them does not depend on it.
		for _, name := range []string{"sharedmail1", "sharedmail2"} {
			_, err := RunWithTestContext(UsersAddCommand, []string{"add", "--email=shared@example.com", "--password=test1234", "--role=guest", name})
			require.NoError(t, err)

			m := entity.FindUserByName(name)
			require.NotNil(t, m)
			t.Cleanup(func() {
				assert.NoError(t, entity.UnscopedDb().Delete(&entity.UserDetails{}, "user_uid = ?", m.UserUID).Error)
				assert.NoError(t, entity.UnscopedDb().Delete(&entity.UserSettings{}, "user_uid = ?", m.UserUID).Error)
				assert.NoError(t, entity.UnscopedDb().Delete(&entity.User{}, "user_uid = ?", m.UserUID).Error)
			})
		}

		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--role=admin", "sharedmail1"})
		require.NoError(t, err)

		m := entity.FindUserByName("sharedmail1")
		require.NotNil(t, m)
		assert.Equal(t, "admin", m.UserRole)
		assert.Equal(t, "shared@example.com", m.UserEmail)

		// Changing the address clears its verification.
		require.NoError(t, entity.UnscopedDb().Model(m).Update("VerifiedAt", entity.TimeStamp()).Error)
		require.True(t, entity.FindUserByName("sharedmail1").EmailVerified())

		_, err = RunWithTestContext(UsersModCommand, []string{"mod", "--email=other@example.com", "sharedmail1"})
		require.NoError(t, err)

		m = entity.FindUserByName("sharedmail1")
		require.NotNil(t, m)
		assert.Equal(t, "other@example.com", m.UserEmail)
		assert.False(t, m.EmailVerified())
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

func TestUsersModCommand_Restore(t *testing.T) {
	requireTestDb(t)
	t.Setenv("PHOTOPRISM_CLI", "")

	t.Run("NoTerminal", func(t *testing.T) {
		newDeletedTestUser(t, "restoremodtty")

		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--name=Changed", "restoremodtty"})

		require.Error(t, err)
		assert.Equal(t, 1, ExitCode(err))
		assert.Contains(t, err.Error(), "--restore")
		assert.True(t, entity.FindUserByName("restoremodtty").IsDeleted())
	})
	t.Run("NonInteractiveEnv", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_CLI", NONINTERACTIVE)
		newDeletedTestUser(t, "restoremodenv")

		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--name=Changed", "restoremodenv"})

		require.Error(t, err)
		assert.Equal(t, 1, ExitCode(err))
		assert.Contains(t, err.Error(), "--restore")
		assert.True(t, entity.FindUserByName("restoremodenv").IsDeleted())
	})
	t.Run("Declined", func(t *testing.T) {
		newDeletedTestUser(t, "restoremodno")
		pipeResetAnswers(t, "n\n")

		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--name=Changed", "restoremodno"})

		assert.EqualError(t, err, "user already exists")

		m := entity.FindUserByName("restoremodno")
		assert.True(t, m.IsDeleted())
		assert.NotEqual(t, "Changed", m.DisplayName)
	})
	t.Run("Confirmed", func(t *testing.T) {
		newDeletedTestUser(t, "restoremodyes")
		pipeResetAnswers(t, "y\n")

		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--name=Changed", "restoremodyes"})

		assert.NoError(t, err)

		m := entity.FindUserByName("restoremodyes")
		assert.False(t, m.IsDeleted())
		assert.Equal(t, "Changed", m.DisplayName)
	})
	t.Run("YesDoesNotRestore", func(t *testing.T) {
		newDeletedTestUser(t, "restoremodyes2")

		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--yes", "--name=Changed", "restoremodyes2"})

		assert.Error(t, err)
		assert.True(t, entity.FindUserByName("restoremodyes2").IsDeleted())
	})
	t.Run("RestoreFlag", func(t *testing.T) {
		newDeletedTestUser(t, "restoremodflag")

		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--restore", "--name=Changed", "restoremodflag"})

		assert.NoError(t, err)
		assert.False(t, entity.FindUserByName("restoremodflag").IsDeleted())
	})
}
