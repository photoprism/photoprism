package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestCientsRemoveCommand(t *testing.T) {
	t.Run("NoConfirmationProvided", func(t *testing.T) {
		// Run command with test context.
		output0, err := RunWithTestContext(ClientsShowCommand, []string{"show", "cs7pvt5h8rw9aaqj"})

		// t.Logf(output0)
		assert.NoError(t, err)
		assert.NotContains(t, output0, "not found")
		assert.Contains(t, output0, "client")

		t.Setenv("PHOTOPRISM_CLI", "")

		// Without a terminal the prompt cannot run, which is a usage error rather than a refusal.
		_, err = RunWithTestContext(ClientsRemoveCommand, []string{"rm", "cs7pvt5h8rw9aaqj"})

		var exit cli.ExitCoder
		require.ErrorAs(t, err, &exit)
		assert.Equal(t, 2, exit.ExitCode())
		assert.Contains(t, err.Error(), "--yes")

		// Run command with test context.
		output2, err := RunWithTestContext(ClientsShowCommand, []string{"show", "cs7pvt5h8rw9aaqj"})

		// t.Logf(output2)
		assert.NoError(t, err)
		assert.NotContains(t, output2, "not found")
		assert.Contains(t, output2, "client")
	})
	t.Run("RemoveClient", func(t *testing.T) {
		restoreAnalyticsSession(t)
		t.Cleanup(func() {
			reopenConnection()
			m := entity.ClientFixtures.Get("analytics")
			assert.NoError(t, m.Restore(), "restore client fixture")
		})

		// Run command with test context.
		output0, err := RunWithTestContext(ClientsShowCommand, []string{"show", "cs7pvt5h8rw9aaqj"})

		// t.Logf(output0)
		assert.NoError(t, err)
		assert.NotContains(t, output0, "not found")
		assert.Contains(t, output0, "client")

		// Run command with test context.
		output, err := RunWithTestContext(ClientsRemoveCommand, []string{"rm", "--yes", "cs7pvt5h8rw9aaqj"})

		// Check command output for plausibility.
		assert.NoError(t, err)
		assert.Empty(t, output)

		// The record is retained so its identifiers stay reserved, and an operator
		// reaching it by id still sees it, with the time it was deleted.
		output2, err := RunWithTestContext(ClientsShowCommand, []string{"show", "cs7pvt5h8rw9aaqj"})

		assert.NoError(t, err)
		assert.Contains(t, output2, "DeletedAt")

		// It no longer appears among the registered clients.
		list, err := RunWithTestContext(ClientsListCommand, []string{"ls", "cs7pvt5h8rw9aaqj"})

		assert.NoError(t, err)
		assert.NotContains(t, list, "cs7pvt5h8rw9aaqj")

		// Listing the deleted ones reports it.
		deleted, err := RunWithTestContext(ClientsListCommand, []string{"ls", "--deleted", "cs7pvt5h8rw9aaqj"})

		assert.NoError(t, err)
		assert.Contains(t, deleted, "cs7pvt5h8rw9aaqj")
	})
	t.Run("NotFound", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(ClientsRemoveCommand, []string{"rm", "--yes", "cs7pvt5h8rw9a000"})

		// Check command output for plausibility.
		assertExitCode(t, err, 3)
		assert.Empty(t, output)
	})
}

// TestClientsRemoveCommand_ForceAlias pins that the deprecated --force flag still confirms the removal.
func TestClientsRemoveCommand_ForceAlias(t *testing.T) {
	t.Setenv("PHOTOPRISM_CLI", "")

	m := entity.NewClient().SetName("ForceAlias").SetScope("metrics")
	require.NoError(t, m.Create())

	_, err := RunWithTestContext(ClientsRemoveCommand, []string{"rm", "-f", m.ClientUID})
	require.NoError(t, err)

	found := entity.FindClient(m.ClientUID)
	require.NotNil(t, found)
	assert.True(t, found.Deleted())
}

func TestClientsRemoveCommand_Purge(t *testing.T) {
	t.Run("ReleasesTheClientUID", func(t *testing.T) {
		m := entity.NewClient().SetName("PurgeMe").SetScope("metrics")

		if err := m.Create(); err != nil {
			t.Fatal(err)
		}

		uid := m.ClientUID

		output, err := RunWithTestContext(ClientsRemoveCommand, []string{"rm", "--yes", "--purge", uid})

		assert.NoError(t, err)
		assert.Empty(t, output)
		assert.Nil(t, entity.FindClientByUID(uid))
	})
	t.Run("PurgesAnAlreadyDeletedClient", func(t *testing.T) {
		m := entity.NewClient().SetName("PurgeDeleted").SetScope("metrics")

		if err := m.Create(); err != nil {
			t.Fatal(err)
		}

		uid := m.ClientUID

		if err := m.Delete(); err != nil {
			t.Fatal(err)
		}

		// An ordinary remove refuses a record that is already retired.
		_, err := RunWithTestContext(ClientsRemoveCommand, []string{"rm", "--yes", uid})
		assertExitCode(t, err, 3)

		_, err = RunWithTestContext(ClientsRemoveCommand, []string{"rm", "--yes", "--purge", uid})
		assert.NoError(t, err)
		assert.Nil(t, entity.FindClientByUID(uid))
	})
	t.Run("RefusesANodeClient", func(t *testing.T) {
		m := entity.NewClient().SetName("pp-purge-node").SetScope("metrics")
		m.NodeUUID = rnd.UUIDv7()

		if err := m.Create(); err != nil {
			t.Fatal(err)
		}

		// Releasing a node UUID also means releasing its database and grants, so that
		// goes through the cluster command instead.
		_, err := RunWithTestContext(ClientsRemoveCommand, []string{"rm", "--yes", "--purge", m.ClientUID})

		assertExitCode(t, err, 2)
		assert.Contains(t, err.Error(), "cluster nodes rm --purge")

		assert.NotNil(t, entity.FindClientByUID(m.ClientUID))
	})
}

func TestClientsModCommand_Restore(t *testing.T) {
	t.Run("ByClientUID", func(t *testing.T) {
		m := entity.NewClient().SetName("RestoreMe").SetScope("metrics")

		if err := m.Create(); err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() {
			reopenConnection()
			assert.NoError(t, m.Purge(), "purge test client")
		})

		uid := m.ClientUID

		if err := m.Delete(); err != nil {
			t.Fatal(err)
		}

		// A deleted client cannot authenticate until it is restored.
		assert.True(t, entity.FindClientByUID(uid).Disabled())

		_, err := RunWithTestContext(ClientsModCommand, []string{"mod", "--restore", uid})
		assert.NoError(t, err)

		if restored := entity.FindClientByUID(uid); assert.NotNil(t, restored) {
			assert.False(t, restored.Deleted())
		}
	})
	t.Run("ByNodeUUID", func(t *testing.T) {
		uuid := rnd.UUIDv7()

		m := entity.NewClient().SetName("pp-restore-node").SetScope("metrics")
		m.NodeUUID = uuid

		if err := m.Create(); err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() {
			reopenConnection()
			assert.NoError(t, m.Purge(), "purge test client")
		})

		uid := m.ClientUID

		if err := m.Delete(); err != nil {
			t.Fatal(err)
		}

		// An operator names a node by the UUID it reports, not by its client uid.
		_, err := RunWithTestContext(ClientsModCommand, []string{"mod", "--restore", uuid})
		assert.NoError(t, err)

		if restored := entity.FindClientByUID(uid); assert.NotNil(t, restored) {
			assert.False(t, restored.Deleted())
		}

		assert.NotNil(t, entity.FindClientByNodeUUID(uuid))
	})
	t.Run("WithoutRestoreFlag", func(t *testing.T) {
		m := entity.NewClient().SetName("RestoreRefused").SetScope("metrics")

		if err := m.Create(); err != nil {
			t.Fatal(err)
		}

		uid := m.ClientUID

		if err := m.Delete(); err != nil {
			t.Fatal(err)
		}

		// Nothing is brought back by accident: the record stays retired.
		_, err := RunWithTestContext(ClientsModCommand, []string{"mod", uid})
		assert.Error(t, err)
		assert.True(t, entity.FindClientByUID(uid).Deleted())
	})
}
