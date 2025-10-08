package cluster

import (
	"github.com/photoprism/photoprism/internal/auth/acl"
)

type NodeRole = string

const (
	RolePortal   = NodeRole(acl.RolePortal)   // A management portal for orchestrating a cluster
	RoleInstance = NodeRole(acl.RoleInstance) // A regular PhotoPrism instance that can join a cluster
	RoleService  = NodeRole(acl.RoleService)  // Other service used within a cluster, e.g. Ollama or Vision API
)
