package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/dsn"
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
	t.Run("ExistingNotASCII", func(t *testing.T) {
		m := entity.NewUser()
		m.UserName = "mod-josé"
		m.UserRole = "guest"
		m.AuthProvider = authn.ProviderLocal.String()

		require.NoError(t, entity.Db().Create(m).Error)
		t.Cleanup(func() {
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.UserDetails{}, "user_uid = ?", m.UserUID).Error)
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.UserSettings{}, "user_uid = ?", m.UserUID).Error)
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.User{}, "user_uid = ?", m.UserUID).Error)
		})

		// Other values change without changing the stored name.
		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--name=José Example", "--role=admin", "--email=jose@example.com", "mod-josé"})
		require.NoError(t, err)

		stored := entity.FindUserByUID(m.UserUID)
		require.NotNil(t, stored)
		assert.Equal(t, "mod-josé", stored.UserName)
		assert.Equal(t, "José Example", stored.DisplayName)
		assert.Equal(t, "admin", stored.UserRole)
		assert.Equal(t, "jose@example.com", stored.UserEmail)

		// A new name must be ASCII.
		_, err = RunWithTestContext(UsersModCommand, []string{"mod", "--username=mod-josé2", "mod-josé"})
		require.ErrorIs(t, err, authn.ErrNotASCII)
		assert.Equal(t, "mod-josé", entity.FindUserByUID(m.UserUID).UserName)

		_, err = RunWithTestContext(UsersModCommand, []string{"mod", "--username=mod-jose2", "mod-josé"})
		require.NoError(t, err)
		assert.Equal(t, "mod-jose2", entity.FindUserByUID(m.UserUID).UserName)
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
	t.Run("RestoreOtherName", func(t *testing.T) {
		m := entity.NewUser()
		m.UserName = "restoremod-josé"
		m.UserRole = "admin"
		require.NoError(t, entity.Db().Create(m).Error)
		t.Cleanup(func() {
			reopenConnection()
			removeTestUser(t, m.UserUID)
		})
		require.NoError(t, m.Delete())

		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--restore", "--name=Changed", "restoremod-jose"})
		require.EqualError(t, err, "user 'restoremod-jose' not found")
		assert.True(t, entity.FindUser(entity.User{UserUID: m.UserUID}).IsDeleted())

		_, err = RunWithTestContext(UsersModCommand, []string{"mod", "--restore", "--name=Changed", m.UserUID})
		require.NoError(t, err)
		restored := entity.FindUser(entity.User{UserUID: m.UserUID})
		assert.False(t, restored.IsDeleted())
		assert.Equal(t, "restoremod-josé", restored.UserName)
	})
	t.Run("RestoreLegacyName", func(t *testing.T) {
		m := entity.NewUser()
		m.UserName = "Restoremod-Legacy"
		m.UserRole = "admin"
		require.NoError(t, entity.Db().Create(m).Error)
		t.Cleanup(func() {
			reopenConnection()
			removeTestUser(t, m.UserUID)
		})
		require.NoError(t, m.Delete())

		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--restore", "restoremod-legacy"})

		// Only MariaDB finds the stored name by its sanitized form.
		if entity.IsDialect(dsn.DriverMySQL) {
			require.NoError(t, err)
			restored := entity.FindUser(entity.User{UserUID: m.UserUID})
			assert.False(t, restored.IsDeleted())
			assert.Equal(t, "restoremod-legacy", restored.UserName)
		} else {
			require.EqualError(t, err, "user 'restoremod-legacy' not found")
		}
	})
	t.Run("RestoreHeld", func(t *testing.T) {
		newDeletedTestUser(t, "restoremodheld")
		deleted := entity.FindUserByName("restoremodheld")
		require.NotNil(t, deleted)
		active := &entity.User{UserName: "restoremodheld", UserRole: "admin", CanLogin: true}
		require.NoError(t, active.Create())
		t.Cleanup(func() {
			reopenConnection()
			removeTestUser(t, active.UserUID)
		})

		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--restore", deleted.UserUID})
		require.EqualError(t, err, "user 'restoremodheld' already exists")
		assert.True(t, entity.FindUser(entity.User{UserUID: deleted.UserUID}).IsDeleted())
	})
}
