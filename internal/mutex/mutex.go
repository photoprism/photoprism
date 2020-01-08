package mutex

import (
	"sync"
)

var Db = sync.Mutex{}

var Worker = Busy{}
