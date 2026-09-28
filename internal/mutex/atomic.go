package mutex

import (
	"sync"
	"sync/atomic"
)

// Restart signals that the application should be restarted,
// e.g. after an update or a config changes.
var Restart = atomic.Bool{}

// TempArchives signals that download archives may exist in the temp directory, so workers can skip
// scanning it. Starts true to cover archives left by a previous process, is set when an archive is
// created, and is cleared by a sweep that finds none left.
var TempArchives = atomic.Bool{}

// UserUploads signals that staged batches may need an expiry scan.
// It starts true for batches left by a previous process and is set when a batch is staged.
var UserUploads = atomic.Bool{}

// UploadBatches coordinates active upload requests with batch expiry.
var UploadBatches sync.RWMutex

// init arms expiry scans for files left by a previous process.
func init() {
	TempArchives.Store(true)
	UserUploads.Store(true)
}
