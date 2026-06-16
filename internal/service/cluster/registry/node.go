package registry

import "github.com/photoprism/photoprism/internal/service/cluster"

// Node represents a registered cluster node (transport DTO inside registry package).
// It embeds the public cluster.Node DTO so we have a single source of truth for fields.
// Additional internal-only metadata is stored alongside the embedded struct.
type Node struct {
	cluster.Node
	ClientSecret string `json:"-"`                   // plaintext only when newly created/rotated in-memory
	RotatedAt    string `json:"RotatedAt,omitempty"` // secret rotation timestamp
	AuthEnabled  bool   `json:"-"`                   // auth client is enabled
	// NameSrc carries the DisplayName provenance for this write (entity.SrcAuto
	// for instance registrations, entity.SrcManual for admin overrides) so Put
	// can apply the same source-priority rule as User.SetDisplayName. It is a
	// write-control field, not serialized.
	NameSrc string `json:"-"`
	// GroupsSrc carries the group-config provenance for this write
	// (entity.ClientGroupsSrcNode for instance registrations,
	// entity.ClientGroupsSrcManual for admin edits); callers that don't manage
	// group config leave it empty. It is a write-control field, not serialized.
	GroupsSrc string `json:"-"`
}

// ensureDatabase returns a writable NodeDatabase, creating one if missing.
func (n *Node) ensureDatabase() *cluster.NodeDatabase {
	if n.Database == nil {
		n.Database = &cluster.NodeDatabase{}
	}

	return n.Database
}
