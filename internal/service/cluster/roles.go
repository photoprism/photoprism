package cluster

import (
	"strings"

	"github.com/photoprism/photoprism/internal/auth/acl"
)

// NodeRole represents the role a node plays within a cluster.
type NodeRole = string

const (
	// RoleInstance represents a regular PhotoPrism instance that can join a cluster.
	RoleInstance = NodeRole(acl.RoleInstance)
	// RolePortal represents a management portal for orchestrating a cluster.
	RolePortal = NodeRole(acl.RolePortal)
	// RoleService represents other services used within a cluster, e.g., Ollama or Vision API.
	RoleService = NodeRole(acl.RoleService)
)

// NormalizeNodeRole maps cluster role aliases to their canonical values.
func NormalizeNodeRole(role string) NodeRole {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "app", RoleInstance:
		return RoleInstance
	case RolePortal:
		return RolePortal
	case RoleService:
		return RoleService
	default:
		return ""
	}
}
