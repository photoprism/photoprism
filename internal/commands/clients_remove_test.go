package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"

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

		// Run command with test context.
		output, err := RunWithTestContext(ClientsRemoveCommand, []string{"rm", "cs7pvt5h8rw9aaqj"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Empty(t, output)

		// Run command with test context.
		output2, err := RunWithTestContext(ClientsShowCommand, []string{"show", "cs7pvt5h8rw9aaqj"})

		// t.Logf(output2)
		assert.NoError(t, err)
		assert.NotContains(t, output2, "not found")
		assert.Contains(t, output2, "client")
	})
	t.Run("RemoveClient", func(t *testing.T) {
		// Run command with test context.
		output0, err := RunWithTestContext(ClientsShowCommand, []string{"show", "cs7pvt5h8rw9aaqj"})

		// t.Logf(output0)
		assert.NoError(t, err)
		assert.NotContains(t, output0, "not found")
		assert.Contains(t, output0, "client")

		// Run command with test context.
		output, err := RunWithTestContext(ClientsRemoveCommand, []string{"rm", "--force", "cs7pvt5h8rw9aaqj"})

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
		output, err := RunWithTestContext(ClientsRemoveCommand, []string{"rm", "--force", "cs7pvt5h8rw9a000"})

		// Check command output for plausibility.
		assert.Error(t, err)
		assert.Empty(t, output)
	})
}

func TestClientsRemoveCommand_Purge(t *testing.T) {
	t.Run("ReleasesTheClientUID", func(t *testing.T) {
		m := entity.NewClient().SetName("PurgeMe").SetScope("metrics")

		if err := m.Create(); err != nil {
			t.Fatal(err)
		}

		uid := m.ClientUID

		output, err := RunWithTestContext(ClientsRemoveCommand, []string{"rm", "--force", "--purge", uid})

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
		_, err := RunWithTestContext(ClientsRemoveCommand, []string{"rm", "--force", uid})
		assert.Error(t, err)

		_, err = RunWithTestContext(ClientsRemoveCommand, []string{"rm", "--force", "--purge", uid})
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
		_, err := RunWithTestContext(ClientsRemoveCommand, []string{"rm", "--force", "--purge", m.ClientUID})

		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "cluster nodes rm --purge")
		}

		assert.NotNil(t, entity.FindClientByUID(m.ClientUID))
	})
}

func TestClientsModCommand_Restore(t *testing.T) {
	t.Run("ByClientUID", func(t *testing.T) {
		m := entity.NewClient().SetName("RestoreMe").SetScope("metrics")

		if err := m.Create(); err != nil {
			t.Fatal(err)
		}

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
