package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/service/cluster"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
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
