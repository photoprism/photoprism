package registry

import "github.com/photoprism/photoprism/internal/service/cluster"

// Node represents a registered cluster node (transport DTO inside registry package).
// It embeds the public cluster.Node DTO so we have a single source of truth for fields.
// Additional internal-only metadata is stored alongside the embedded struct.
type Node struct {
	cluster.Node
	ClientSecret string `json:"-"`                   // plaintext only when newly created/rotated in-memory
	RotatedAt    string `json:"rotatedAt,omitempty"` // secret rotation timestamp
}

// ensureDatabase returns a writable NodeDatabase, creating one if missing.
func (n *Node) ensureDatabase() *cluster.NodeDatabase {
	if n.Node.Database == nil {
		n.Node.Database = &cluster.NodeDatabase{}
	}
	return n.Node.Database
}
