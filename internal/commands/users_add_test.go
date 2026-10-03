package commands

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/txt"
)

func TestUsersAddCommand(t *testing.T) {
	t.Run("AddUserThatAlreadyExists", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(UsersAddCommand, []string{"add", "--name=Alice", "--email=jane@test.de", "--password=test1234", "--role=admin", "alice"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)

	})
	t.Run("AddDeletedUser", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_CLI", "")

		// Restoring the deleted account needs a confirmation, which cannot be asked without a terminal.
		_, err := RunWithTestContext(UsersAddCommand, []string{"add", "--name=deleted", "--password=test1234", "deleted"})

		assert.Equal(t, 1, ExitCode(err))
		assert.True(t, entity.FindUserByName("deleted").IsDeleted())
	})
	t.Run("AddUsernameMissing", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(UsersAddCommand, []string{"add", "--name=noname", "--password=test1234", "/##"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)
	})
}

// newDeletedTestUser creates and deletes a user account, and removes it for good when the test ends.
func newDeletedTestUser(t *testing.T, name string) {
	t.Helper()

	m := &entity.User{UserName: name, UserRole: "admin", CanLogin: true}
	require.NoError(t, m.Create())
	t.Cleanup(func() {
		reopenConnection()
		removeTestUser(t, m.UserUID)
	})
	require.NoError(t, m.Delete())
	require.True(t, entity.FindUserByName(name).IsDeleted())
}

func TestUsersAddCommand_Restore(t *testing.T) {
	requireTestDb(t)
	t.Setenv("PHOTOPRISM_CLI", "")

	t.Run("NoTerminal", func(t *testing.T) {
		newDeletedTestUser(t, "restoreaddtty")

		_, err := RunWithTestContext(UsersAddCommand, []string{"add", "--password=test1234", "restoreaddtty"})

		require.Error(t, err)
		assert.Equal(t, 1, ExitCode(err))
		assert.Contains(t, err.Error(), "--restore")
		assert.True(t, entity.FindUserByName("restoreaddtty").IsDeleted())
	})
	t.Run("NonInteractiveEnv", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_CLI", NONINTERACTIVE)
		newDeletedTestUser(t, "restoreaddenv")

		_, err := RunWithTestContext(UsersAddCommand, []string{"add", "--password=test1234", "restoreaddenv"})

		require.Error(t, err)
		assert.Equal(t, 1, ExitCode(err))
		assert.Contains(t, err.Error(), "--restore")
		assert.True(t, entity.FindUserByName("restoreaddenv").IsDeleted())
	})
	t.Run("Declined", func(t *testing.T) {
		newDeletedTestUser(t, "restoreaddno")
		pipeResetAnswers(t, "n\n")

		_, err := RunWithTestContext(UsersAddCommand, []string{"add", "--password=test1234", "restoreaddno"})

		assert.ErrorIs(t, err, authn.ErrAccountAlreadyExists)
		assert.True(t, entity.FindUserByName("restoreaddno").IsDeleted())
	})
	t.Run("Confirmed", func(t *testing.T) {
		newDeletedTestUser(t, "restoreaddyes")
		pipeResetAnswers(t, "y\n")

		_, err := RunWithTestContext(UsersAddCommand, []string{"add", "--password=test1234", "restoreaddyes"})

		assert.NoError(t, err)
		assert.False(t, entity.FindUserByName("restoreaddyes").IsDeleted())
	})
	t.Run("RestoreFlag", func(t *testing.T) {
		newDeletedTestUser(t, "restoreaddflag")

		_, err := RunWithTestContext(UsersAddCommand, []string{"add", "--restore", "--password=test1234", "restoreaddflag"})

		assert.NoError(t, err)
		assert.False(t, entity.FindUserByName("restoreaddflag").IsDeleted())
	})
	t.Run("RejectedPassword", func(t *testing.T) {
		newDeletedTestUser(t, "restoreaddlong")

		// Too long for any configuration, since the test config runs in public mode without a minimum length.
		_, err := RunWithTestContext(UsersAddCommand, []string{"add", "--restore", "--password=" + strings.Repeat("a", txt.ClipPassword+1), "restoreaddlong"})

		assert.ErrorContains(t, err, "password must have less than")
		assert.True(t, entity.FindUserByName("restoreaddlong").IsDeleted())
	})
	t.Run("YesDoesNotRestore", func(t *testing.T) {
		newDeletedTestUser(t, "restoreaddyes2")

		_, err := RunWithTestContext(UsersAddCommand, []string{"add", "--yes", "--password=test1234", "restoreaddyes2"})

		assert.Error(t, err)
		assert.True(t, entity.FindUserByName("restoreaddyes2").IsDeleted())
	})
	t.Run("NotDeleted", func(t *testing.T) {
		_, err := RunWithTestContext(UsersAddCommand, []string{"add", "--restore", "--password=test1234", "alice"})

		assert.ErrorIs(t, err, authn.ErrAccountAlreadyExists)
		assert.False(t, entity.FindUserByName("alice").IsDeleted())
	})
}

// removeTestUser deletes a user account created by a test, including its details, settings and password.
func removeTestUser(t *testing.T, uid string) {
	t.Helper()

	db := entity.UnscopedDb()
	assert.NoError(t, db.Delete(entity.UserDetails{}, "user_uid = ?", uid).Error)
	assert.NoError(t, db.Delete(entity.UserSettings{}, "user_uid = ?", uid).Error)
	assert.NoError(t, db.Delete(entity.Password{}, "uid = ?", uid).Error)
	assert.NoError(t, db.Delete(entity.User{}, "user_uid = ?", uid).Error)
}
