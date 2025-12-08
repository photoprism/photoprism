/*
Package registry defines and implements the cluster node registry abstraction.

The default implementation stores nodes in the Portal's OAuth client table
and exposes CRUD-style operations to create, lookup, list, delete, and rotate
secrets for nodes. Helper mappers convert internal Node records to API/CLI
response DTOs while redacting sensitive fields.

Copyright (c) 2018 - 2025 PhotoPrism UG. All rights reserved.

	This program is free software: you can redistribute it and/or modify
	it under Version 3 of the GNU Affero General Public License (the "AGPL"):
	<https://docs.photoprism.app/license/agpl>

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	The AGPL is supplemented by our Trademark and Brand Guidelines,
	which describe how our Brand Assets may be used:
	<https://www.photoprism.app/trademark>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>
*/
package registry

import (
	"os"

	"github.com/photoprism/photoprism/internal/service/cluster"
)

// Registry abstracts cluster node persistence so we can back it with auth_clients.
// Implementations should be Portal-local and enforce no cross-process locking here.
type Registry interface {
	Put(n *Node) error // Core CRUD.
	Get(uuid string) (*Node, error)
	FindByName(name string) (*Node, error)
	List() ([]Node, error)
	Delete(uuid string) error                  // Typical case: a single record to delete or update.
	DeleteAllByUUID(uuid string) error         // DeleteAllByUUID removes all client records that share the given UUID.
	RotateSecret(uuid string) (*Node, error)   // Secret rotation by UUID (primary identifier).
	FindByNodeUUID(uuid string) (*Node, error) // UUID-first helpers (primary identifier for nodes).
	FindByClientID(clientID string) (*Node, error)
	GetClusterNodeByUUID(uuid string, opts NodeOpts) (cluster.Node, error)
}

// ErrNotFound is returned when a node cannot be found.
var ErrNotFound = os.ErrNotExist
