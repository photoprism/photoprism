package registry

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/service/cluster"
)

// NodeOpts controls which optional fields get included in responses.
type NodeOpts struct {
	IncludeClientID     bool
	IncludeAdvertiseUrl bool
	IncludeDatabase     bool
	IncludeAccessRules  bool
}

// NodeOptsForSession returns the default exposure policy for a session.
// Admin users see the OAuth client identifier, AdvertiseUrl, DB metadata, and group-based
// access rules; others get a redacted view.
func NodeOptsForSession(s *entity.Session) NodeOpts {
	if s != nil && s.GetUser() != nil && s.GetUser().IsAdmin() {
		return NodeOpts{IncludeClientID: true, IncludeAdvertiseUrl: true, IncludeDatabase: true, IncludeAccessRules: true}
	}

	return NodeOpts{}
}

// NodeOptsForSelf returns the exposure policy for a node's own registration response.
// A node stores the client identifier it is issued, so its own record reports it while
// other nodes stay redacted.
func NodeOptsForSelf() NodeOpts {
	return NodeOpts{IncludeClientID: true}
}

// NodeOptsForOperator returns the exposure policy for local operator tooling, which runs
// against the registry rather than through a session. Group access rules are omitted because
// the dedicated access commands report them.
func NodeOptsForOperator() NodeOpts {
	return NodeOpts{IncludeClientID: true, IncludeAdvertiseUrl: true, IncludeDatabase: true}
}

// BuildClusterNode builds a cluster.Node DTO from a registry.Node with redaction according to opts.
func BuildClusterNode(n Node, opts NodeOpts) cluster.Node {
	out := n.Node

	if !opts.IncludeClientID {
		out.ClientID = ""
	}

	if !opts.IncludeAdvertiseUrl {
		out.AdvertiseUrl = ""
	}

	if !opts.IncludeDatabase {
		out.Database = nil
	}

	if !opts.IncludeAccessRules {
		out.AllowGroups = nil
		out.AllowGroupRoles = nil
		out.GroupsFullView = nil
		out.GroupsSrc = ""
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
