package commands

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// newExactTestUser inserts an account without the checks of Create, as one stored earlier, deletes it
// if deleted is true, and removes it after the test.
func newExactTestUser(t *testing.T, name string, deleted bool) *entity.User {
	t.Helper()

	m := entity.NewUser()
	m.UserName = name
	m.UserRole = "admin"
	m.AuthProvider = authn.ProviderLocal.String()
	m.CanLogin = true

	require.NoError(t, entity.Db().Create(m).Error)
	t.Cleanup(func() {
		reopenConnection()
		removeTestUser(t, m.UserUID)
	})

	if deleted {
		require.NoError(t, m.Delete())
	}

	return m
}

func TestUsersCommands_ExactName(t *testing.T) {
	requireTestDb(t)
	t.Setenv("PHOTOPRISM_CLI", "")

	t.Run("ModOtherName", func(t *testing.T) {
		other := newExactTestUser(t, "exact-mod-josé", false)
		m := newExactTestUser(t, "exact-mod-jose", false)

		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--name=Changed", "exact-mod-jose"})
		require.NoError(t, err)

		assert.Equal(t, "Changed", entity.FindUserByUID(m.UserUID).DisplayName)
		assert.NotEqual(t, "Changed", entity.FindUserByUID(other.UserUID).DisplayName)
	})
	t.Run("ModDeletedLowerID", func(t *testing.T) {
		deleted := newExactTestUser(t, "exact-mod-dup", true)
		m := newExactTestUser(t, "exact-mod-dup", false)

		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--name=Changed", "exact-mod-dup"})
		require.NoError(t, err)

		assert.Equal(t, "Changed", entity.FindUserByUID(m.UserUID).DisplayName)
		assert.True(t, entity.FindUser(entity.User{UserUID: deleted.UserUID}).IsDeleted())
	})
	t.Run("ModRestoreActive", func(t *testing.T) {
		m := newExactTestUser(t, "exact-mod-active", false)
		logs := captureCommandLog(t)

		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--restore", "--name=Changed", "exact-mod-active"})
		require.NoError(t, err)

		assert.Equal(t, "Changed", entity.FindUserByUID(m.UserUID).DisplayName)
		assert.Contains(t, logMessages(logs, logrus.InfoLevel), "user 'exact-mod-active' is not deleted, nothing to restore")
		assert.Contains(t, logMessages(logs, logrus.InfoLevel), "user 'exact-mod-active' ("+m.UserUID+") has been updated")
	})
	t.Run("ModRestoreBesideOtherName", func(t *testing.T) {
		deleted := newExactTestUser(t, "exact-restore-jose", true)
		newExactTestUser(t, "exact-restore-josé", false)

		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--restore", deleted.UserUID})

		// MariaDB compares names that differ only by accents as equal.
		if entity.IsDialect(dsn.DriverMySQL) {
			require.EqualError(t, err, "user 'exact-restore-jose' already exists")
			assert.True(t, entity.FindUser(entity.User{UserUID: deleted.UserUID}).IsDeleted())
		} else {
			require.NoError(t, err)
			assert.False(t, entity.FindUser(entity.User{UserUID: deleted.UserUID}).IsDeleted())
		}
	})
	t.Run("ShowOtherName", func(t *testing.T) {
		other := newExactTestUser(t, "exact-show-josé", false)
		m := newExactTestUser(t, "exact-show-jose", false)

		output, err := RunWithTestContext(UsersShowCommand, []string{"show", "exact-show-jose"})
		require.NoError(t, err)

		assert.Contains(t, output, m.UserUID)
		assert.NotContains(t, output, other.UserUID)
	})
	t.Run("ShowNotFound", func(t *testing.T) {
		newExactTestUser(t, "exact-none-josé", false)

		_, err := RunWithTestContext(UsersShowCommand, []string{"show", "exact-none-jose"})
		require.Error(t, err)
	})
	t.Run("RemoveOtherName", func(t *testing.T) {
		other := newExactTestUser(t, "exact-rm-josé", false)
		m := newExactTestUser(t, "exact-rm-jose", false)

		_, err := RunWithTestContext(UsersRemoveCommand, []string{"rm", "--yes", "exact-rm-jose"})
		require.NoError(t, err)

		assert.True(t, entity.FindUser(entity.User{UserUID: m.UserUID}).IsDeleted())
		assert.False(t, entity.FindUser(entity.User{UserUID: other.UserUID}).IsDeleted())
	})
	t.Run("PasswdOtherName", func(t *testing.T) {
		other := newExactTestUser(t, "exact-pw-josé", false)
		m := newExactTestUser(t, "exact-pw-jose", false)
		require.NoError(t, other.SetPassword("Other-Password-1"))
		require.NoError(t, m.SetPassword("Exact-Password-1"))

		_, err := RunWithTestContext(PasswdCommand, []string{"passwd", "--rm", "exact-pw-jose"})
		require.NoError(t, err)

		assert.Nil(t, entity.FindPassword(m.UserUID))
		assert.NotNil(t, entity.FindPassword(other.UserUID))
	})
	t.Run("ModRestoreHeldPrompt", func(t *testing.T) {
		deleted := newExactTestUser(t, "exact-held", true)
		newExactTestUser(t, "exact-held", false)

		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--name=Changed", deleted.UserUID})
		require.EqualError(t, err, "user 'exact-held' already exists")
		assert.True(t, entity.FindUser(entity.User{UserUID: deleted.UserUID}).IsDeleted())
	})
	t.Run("AuthAddDeletedLowerID", func(t *testing.T) {
		deleted := newExactTestUser(t, "exact-auth-dup", true)
		m := newExactTestUser(t, "exact-auth-dup", false)
		t.Cleanup(func() {
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.Session{}, "user_uid IN (?)", []string{m.UserUID, deleted.UserUID}).Error)
		})

		_, err := RunWithTestContext(AuthAddCommand, []string{"add", "--scope=metrics", "--name=exact", "exact-auth-dup"})
		require.NoError(t, err)

		var count, deletedCount int64
		require.NoError(t, entity.UnscopedDb().Model(&entity.Session{}).Where("user_uid = ?", m.UserUID).Count(&count).Error)
		require.NoError(t, entity.UnscopedDb().Model(&entity.Session{}).Where("user_uid = ?", deleted.UserUID).Count(&deletedCount).Error)
		assert.Equal(t, int64(1), count)
		assert.Equal(t, int64(0), deletedCount)
	})
	t.Run("AuthAddDeletedOnly", func(t *testing.T) {
		deleted := newExactTestUser(t, "exact-auth-deleted", true)

		_, err := RunWithTestContext(AuthAddCommand, []string{"add", "--scope=metrics", "--name=exact", "exact-auth-deleted"})
		require.Error(t, err)
		assert.Equal(t, 3, ExitCode(err))

		var count int64
		require.NoError(t, entity.UnscopedDb().Model(&entity.Session{}).Where("user_uid = ?", deleted.UserUID).Count(&count).Error)
		assert.Equal(t, int64(0), count)
	})
	t.Run("ClientsAddUnknownUID", func(t *testing.T) {
		deleted := newExactTestUser(t, "exact-client-uid-deleted", true)

		for _, uid := range []string{rnd.GenerateUID(entity.UserUID), deleted.UserUID} {
			_, err := RunWithTestContext(ClientsAddCommand, []string{"add", "--name=Exact UID Client", "--scope=metrics", uid})
			require.EqualError(t, err, "user "+uid+" not found")
		}

		var count int64
		require.NoError(t, entity.UnscopedDb().Model(&entity.Client{}).Where("client_name = ?", "Exact UID Client").Count(&count).Error)
		assert.Equal(t, int64(0), count)
	})
	t.Run("ClientsAddOtherName", func(t *testing.T) {
		other := newExactTestUser(t, "exact-client-josé", false)
		m := newExactTestUser(t, "exact-client-jose", false)
		t.Cleanup(func() {
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.Client{}, "client_name = ?", "Exact Client").Error)
		})

		_, err := RunWithTestContext(ClientsAddCommand, []string{"add", "--name=Exact Client", "--scope=metrics", "exact-client-jose"})
		require.NoError(t, err)

		var client entity.Client
		require.NoError(t, entity.UnscopedDb().Where("client_name = ?", "Exact Client").First(&client).Error)
		assert.Equal(t, m.UserUID, client.UserUID)
		assert.NotEqual(t, other.UserUID, client.UserUID)
	})
	t.Run("ClientsAddDeletedLowerID", func(t *testing.T) {
		newExactTestUser(t, "exact-client-dup", true)
		m := newExactTestUser(t, "exact-client-dup", false)
		t.Cleanup(func() {
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.Client{}, "client_name = ?", "Exact Dup Client").Error)
		})

		_, err := RunWithTestContext(ClientsAddCommand, []string{"add", "--name=Exact Dup Client", "--scope=metrics", "exact-client-dup"})
		require.NoError(t, err)

		var client entity.Client
		require.NoError(t, entity.UnscopedDb().Where("client_name = ?", "Exact Dup Client").First(&client).Error)
		assert.Equal(t, m.UserUID, client.UserUID)
	})
	t.Run("ClientsAddDeletedOnly", func(t *testing.T) {
		newExactTestUser(t, "exact-client-deleted", true)

		_, err := RunWithTestContext(ClientsAddCommand, []string{"add", "--name=Exact Deleted Client", "--scope=metrics", "exact-client-deleted"})
		require.EqualError(t, err, "user 'exact-client-deleted' not found")
	})
	t.Run("ClientsAddUnknown", func(t *testing.T) {
		newExactTestUser(t, "exact-client-unknown-josé", false)

		_, err := RunWithTestContext(ClientsAddCommand, []string{"add", "--name=Exact Unknown Client", "--scope=metrics", "exact-client-unknown-jose"})
		require.EqualError(t, err, "user 'exact-client-unknown-jose' not found")

		var count int64
		require.NoError(t, entity.UnscopedDb().Model(&entity.Client{}).Where("client_name = ?", "Exact Unknown Client").Count(&count).Error)
		assert.Equal(t, int64(0), count)
	})
}
