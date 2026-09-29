package mutex

import (
	"sync"
	"sync/atomic"
)

// UploadBatches coordinates active upload requests with batch expiry.
var UploadBatches sync.RWMutex

// UploadRequests changes whenever an upload or processing request begins or ends its batch work, so
// batch expiry can tell whether a batch may have changed since it was found to be expired.
var UploadRequests atomic.Uint64

// unlockUploadBatches releases a request's shared lifecycle lock; tests replace it to observe the order.
var unlockUploadBatches = UploadBatches.RUnlock

// BeginUploadRequest holds the batch lifecycle lock shared and records the request.
func BeginUploadRequest() {
	UploadBatches.RLock()
	UploadRequests.Add(1)
}

// EndUploadRequest records the end of a request and releases its shared lifecycle lock. The record
// comes first, so expiry that acquires the lock afterwards sees that the request ran after its scan.
func EndUploadRequest() {
	UploadRequests.Add(1)
	unlockUploadBatches()
}
