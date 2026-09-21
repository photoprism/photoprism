package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/service/cluster"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestNodeDatabase(t *testing.T) {
	withDatabase := func(name, user string) *reg.Node {
		n := &reg.Node{Node: cluster.Node{UUID: "01997832-3b75-77e4-8b7d-b683aa405f3b", Name: "pp-node"}}
		n.Database = &cluster.NodeDatabase{Name: name, User: user}

		return n
	}

	t.Run("Name", func(t *testing.T) {
		// The identifier names the database, so it is not released while that database
		// holds data. This is what the release is checked against.
		assert.Equal(t, "cluster_dabcdefghij", nodeDatabase(withDatabase("cluster_dabcdefghij", "cluster_uabcdefghij")))
	})
	t.Run("UserAlone", func(t *testing.T) {
		assert.Equal(t, "cluster_uabcdefghij", nodeDatabase(withDatabase("", "cluster_uabcdefghij")))
	})
	t.Run("NoneProvisioned", func(t *testing.T) {
		assert.Empty(t, nodeDatabase(withDatabase("", "")))
		assert.Empty(t, nodeDatabase(&reg.Node{Node: cluster.Node{UUID: "01997832-3b75-77e4-8b7d-b683aa405f3b"}}))
	})
	t.Run("Nil", func(t *testing.T) {
		assert.Empty(t, nodeDatabase(nil))
	})
}

// TestReleaseBlockedBy covers the decision a release makes. The record an ordinary lookup
// resolves is not the only one a release removes, so a database named by any of them counts.
func TestReleaseBlockedBy(t *testing.T) {
	t.Run("SiblingRecordBlocks", func(t *testing.T) {
		// The resolved record names no database and is not deleting one, so the sibling's
		// database is what the release would leave behind.
		assert.Equal(t, "cluster_dsibling01", releaseBlockedBy([]string{"cluster_dsibling01"}, ""))
	})
	t.Run("SiblingBlocksEvenWhileCleaningAnother", func(t *testing.T) {
		assert.Equal(t, "cluster_dsibling01", releaseBlockedBy([]string{"cluster_dresolved1", "cluster_dsibling01"}, "cluster_dresolved1"))
	})
	t.Run("TheOneBeingDeletedDoesNotBlock", func(t *testing.T) {
		assert.Empty(t, releaseBlockedBy([]string{"cluster_dresolved1"}, "cluster_dresolved1"))
	})
	t.Run("NoneRecorded", func(t *testing.T) {
		assert.Empty(t, releaseBlockedBy(nil, ""))
		assert.Empty(t, releaseBlockedBy([]string{""}, ""))
	})
}

// TestClusterNodesRemoveCommand_Purge checks release decisions through the Portal command.
func TestClusterNodesRemoveCommand_Purge(t *testing.T) {
	conf := get.Config()
	previousEdition, previousRole := conf.Options().Edition, conf.Options().NodeRole
	conf.Options().Edition, conf.Options().NodeRole = config.Portal, cluster.RolePortal
	t.Cleanup(func() {
		conf.Options().Edition, conf.Options().NodeRole = previousEdition, previousRole
	})

	for _, tc := range []struct {
		name     string
		database bool
		dryRun   bool
	}{
		{name: "SiblingDatabase", database: true},
		{name: "DryRun", database: true, dryRun: true},
		{name: "Unprovisioned"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			uuid := rnd.UUIDv7()
			retired := entity.NewClient().SetName("pp-retired-" + rnd.GenerateUID(entity.ClientUID)).SetRole(cluster.RoleInstance)
			retired.NodeUUID = uuid
			if tc.database {
				retired.SetData(&entity.ClientData{Database: &entity.ClientDatabase{Name: "cluster_dtestmetadata", User: "cluster_utestmetadata", Driver: "mysql"}})
			}
			require.NoError(t, retired.Create())
			t.Cleanup(func() { assert.NoError(t, retired.Purge()) })
			require.NoError(t, retired.Delete())

			current := entity.NewClient().SetName("pp-current-" + rnd.GenerateUID(entity.ClientUID)).SetRole(cluster.RoleInstance)
			current.NodeUUID = uuid
			require.NoError(t, current.Create())
			t.Cleanup(func() { assert.NoError(t, current.Purge()) })
			require.Len(t, entity.FindClientsByNodeUUID(uuid), 2)

			args := []string{"rm", "--purge", "--yes"}
			if tc.dryRun {
				args = append(args, "--dry-run")
			}
			args = append(args, uuid)

			_, err := RunWithTestContext(ClusterNodesRemoveCommand, args)
			if tc.database && !tc.dryRun {
				var exitErr cli.ExitCoder
				require.ErrorAs(t, err, &exitErr)
				assert.Equal(t, 2, exitErr.ExitCode())
				assert.Contains(t, err.Error(), "cluster_dtestmetadata")
			} else {
				require.NoError(t, err)
			}

			if tc.database || tc.dryRun {
				require.Len(t, entity.FindClientsByNodeUUID(uuid), 2)
				storedRetired := entity.FindClientByUID(retired.ClientUID)
				storedCurrent := entity.FindClientByUID(current.ClientUID)
				require.NotNil(t, storedRetired)
				require.NotNil(t, storedCurrent)
				assert.True(t, storedRetired.Deleted())
				assert.False(t, storedCurrent.Deleted())
			} else {
				assert.Empty(t, entity.FindClientsByNodeUUID(uuid))
			}
		})
	}
}
