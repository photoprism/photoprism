package mutex

import (
	"sync"
)

var (
	Db          = sync.Mutex{}
	Index       = sync.Mutex{}
	People      = Busy{}
	MainWorker  = Busy{}
	SyncWorker  = Busy{}
	ShareWorker = Busy{}
	MetaWorker  = Busy{}
	FacesWorker = Busy{}
)

// WorkersBusy returns true if any worker is busy.
func WorkersBusy() bool {
	return MainWorker.Busy() || SyncWorker.Busy() || ShareWorker.Busy() || MetaWorker.Busy() || FacesWorker.Busy()
}
