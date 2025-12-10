package cluster

import (
	"github.com/photoprism/photoprism/internal/auth/acl"
)

// NodeRole represents the role a node plays within a cluster.
type NodeRole = string

const (
	// RoleApp represents a regular PhotoPrism app node that can join a cluster.
	RoleApp = NodeRole(acl.RoleApp)
	// RolePortal represents a management portal for orchestrating a cluster.
	RolePortal = NodeRole(acl.RolePortal)
	// RoleService represents other services used within a cluster, e.g., Ollama or Vision API.
	RoleService = NodeRole(acl.RoleService)
)
