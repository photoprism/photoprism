package mutex

import (
	"sync"
)

var (
	Db     = sync.Mutex{}
	Worker = Busy{}
	Sync   = Busy{}
	Share  = Busy{}
)
