package registry

import "os"

// Registry abstracts cluster node persistence so we can back it with auth_clients.
// Implementations should be Portal-local and enforce no cross-process locking here.
type Registry interface {
	Put(n *Node) error
	Get(id string) (*Node, error)
	FindByName(name string) (*Node, error)
	List() ([]Node, error)
	Delete(id string) error
	RotateSecret(id string) (*Node, error)
}

// ErrNotFound is returned when a node cannot be found.
var ErrNotFound = os.ErrNotExist
