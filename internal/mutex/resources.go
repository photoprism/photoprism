package mutex

import (
	"sync"
)

// Shared resources.
var (
	Db    = sync.Mutex{}
	Index = sync.Mutex{}
)
