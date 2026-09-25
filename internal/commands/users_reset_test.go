package commands

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestUsersResetCommand(t *testing.T) {
	c := resetConfigAndOpenDB()
	// reset as this test removes all users
	defer resetConfigAndDB()

	t.Run("NotConfirmed", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_CLI", "")

		// Run command with test context.
		output0, err := RunWithTestContext(UsersListCommand, []string{"ls"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output0, "alice")
		assert.Contains(t, output0, "bob")

		// Without a terminal the prompt cannot run, which is a usage error rather than a refusal.
		_, err = RunWithTestContext(UsersResetCommand, []string{"reset"})
		assertExitCode(t, err, 2)
		assert.Contains(t, err.Error(), "--yes")

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

		userClients := int64(0)
		require.NoError(t, c.Db().Unscoped().Model(&entity.Client{}).Where("user_uid <> ''").Count(&userClients).Error)
		require.Greater(t, userClients, int64(0))

		aliceUID := entity.UserFixtures.Get("alice").UserUID
		require.NotNil(t, entity.FindPassword(aliceUID))
		_, clientSecrets := countResetTestPasswords(t, c.Db())

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

		// The client applications registered to users are deleted with the accounts.
		remaining, otherClients := int64(1), int64(0)
		require.NoError(t, c.Db().Unscoped().Model(&entity.Client{}).Where("user_uid <> ''").Count(&remaining).Error)
		require.NoError(t, c.Db().Model(&entity.Client{}).Where("user_uid = ''").Count(&otherClients).Error)
		assert.Equal(t, int64(0), remaining)
		assert.Greater(t, otherClients, int64(0))
		assert.Contains(t, buffer.String(), fmt.Sprintf("deleted %d client applications", userClients))

		// Account passwords are removed with the accounts, while the other clients keep their secrets.
		assert.Nil(t, entity.FindPassword(aliceUID))

		var secretsAfter int64
		require.NoError(t, c.Db().Model(&entity.Password{}).Where("uid IN (SELECT client_uid FROM auth_clients)").Count(&secretsAfter).Error)
		assert.Greater(t, secretsAfter, int64(0))
		assert.LessOrEqual(t, secretsAfter, clientSecrets)
	})
}

// countResetTestPasswords returns how many stored passwords belong to user accounts and to clients.
func countResetTestPasswords(t *testing.T, db *gorm.DB) (users, clients int64) {
	t.Helper()

	require.NoError(t, db.Model(&entity.Password{}).Where("uid IN (SELECT user_uid FROM auth_users)").Count(&users).Error)
	require.NoError(t, db.Model(&entity.Password{}).Where("uid IN (SELECT client_uid FROM auth_clients)").Count(&clients).Error)

	return users, clients
}

func TestDeleteUserPasswords(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		c := resetConfigAndOpenDB()
		t.Cleanup(func() { resetConfigAndDB() })

		// Rolled back, so the fixtures stay in place for the tests that follow.
		tx := c.Db().Begin()
		require.NoError(t, tx.Error)
		defer tx.Rollback()

		users, clients := countResetTestPasswords(t, tx)
		require.Greater(t, users, int64(0))
		require.Greater(t, clients, int64(0))

		deleted, err := DeleteUserPasswords(tx)

		require.NoError(t, err)
		assert.Equal(t, users, deleted)

		usersAfter, clientsAfter := countResetTestPasswords(t, tx)
		assert.Zero(t, usersAfter)
		assert.Equal(t, clients, clientsAfter, "client secrets must be kept")
	})
	t.Run("NoDatabase", func(t *testing.T) {
		deleted, err := DeleteUserPasswords(nil)

		assert.Error(t, err)
		assert.Zero(t, deleted)
	})
}

func TestLogDeleteUserPasswords(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		c := resetConfigAndOpenDB()
		t.Cleanup(func() { resetConfigAndDB() })

		tx := c.Db().Begin()
		require.NoError(t, tx.Error)
		defer tx.Rollback()

		buffer := bytes.Buffer{}
		log.SetOutput(&buffer)
		defer log.SetOutput(os.Stdout)

		require.NoError(t, LogDeleteUserPasswords(tx))
		assert.Contains(t, buffer.String(), "account password")
	})
	t.Run("NoDatabase", func(t *testing.T) {
		assert.Error(t, LogDeleteUserPasswords(nil))
	})
}

func TestLogDeleteUserClients(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		c := resetConfigAndOpenDB()
		t.Cleanup(func() { resetConfigAndDB() })

		tx := c.Db().Begin()
		require.NoError(t, tx.Error)
		defer tx.Rollback()

		buffer := bytes.Buffer{}
		log.SetOutput(&buffer)
		defer log.SetOutput(os.Stdout)

		require.NoError(t, LogDeleteUserClients(tx))
		assert.Contains(t, buffer.String(), "deleted 2 client applications")
	})
	t.Run("NoDatabase", func(t *testing.T) {
		assert.Error(t, LogDeleteUserClients(nil))
	})
}

func TestDeleteUserClients(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		c := resetConfigAndOpenDB()
		t.Cleanup(func() { resetConfigAndDB() })

		// Rolled back, so the client fixtures stay in place for the tests that follow.
		tx := c.Db().Begin()
		require.NoError(t, tx.Error)
		defer tx.Rollback()

		// A soft-deleted client registered to a user is deleted as well.
		require.NoError(t, tx.Where("client_uid = ?", entity.ClientFixtures.Get("alice").ClientUID).Delete(&entity.Client{}).Error)

		deleted, err := DeleteUserClients(tx)

		require.NoError(t, err)
		assert.Equal(t, int64(2), deleted)

		remaining, others := int64(1), int64(0)
		require.NoError(t, tx.Unscoped().Model(&entity.Client{}).Where("user_uid <> ''").Count(&remaining).Error)
		require.NoError(t, tx.Model(&entity.Client{}).Where("user_uid = ''").Count(&others).Error)
		assert.Equal(t, int64(0), remaining)
		assert.Greater(t, others, int64(0))
	})
	t.Run("NoDatabase", func(t *testing.T) {
		deleted, err := DeleteUserClients(nil)

		assert.Error(t, err)
		assert.Equal(t, int64(0), deleted)
	})
}
