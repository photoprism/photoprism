package registry

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/service/cluster"
)

// NodeOpts controls which optional fields get included in responses.
type NodeOpts struct {
	IncludeAdvertiseUrl bool
	IncludeDatabase     bool
}

// NodeOptsForSession returns the default exposure policy for a session.
// Admin users see advertiseUrl and DB metadata; others get a redacted view.
func NodeOptsForSession(s *entity.Session) NodeOpts {
	if s != nil && s.GetUser() != nil && s.GetUser().IsAdmin() {
		return NodeOpts{IncludeAdvertiseUrl: true, IncludeDatabase: true}
	}

	return NodeOpts{}
}

// BuildClusterNode builds a cluster.Node DTO from a registry.Node with redaction according to opts.
func BuildClusterNode(n Node, opts NodeOpts) cluster.Node {
	out := n.Node

	if !opts.IncludeAdvertiseUrl {
		out.AdvertiseUrl = ""
	}

	if !opts.IncludeDatabase {
		out.Database = nil
	}

	return out
}

// BuildClusterNodes creates a cluster node slice from the given registry node slice.
func BuildClusterNodes(list []Node, opts NodeOpts) []cluster.Node {
	if len(list) == 0 {
		return []cluster.Node{}
	}

	out := make([]cluster.Node, 0, len(list))

	for _, n := range list {
		out = append(out, BuildClusterNode(n, opts))
	}

	return out
}
