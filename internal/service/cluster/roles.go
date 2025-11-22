package cluster

import (
	"github.com/photoprism/photoprism/internal/auth/acl"
)

type NodeRole = string

const (
	RoleApp     = NodeRole(acl.RoleApp)     // A regular PhotoPrism app node that can join a cluster
	RolePortal  = NodeRole(acl.RolePortal)  // A management portal for orchestrating a cluster
	RoleService = NodeRole(acl.RoleService) // Other service used within a cluster, e.g. Ollama or Vision API
)
