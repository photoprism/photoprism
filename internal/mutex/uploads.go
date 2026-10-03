package mutex

import (
	"hash/maphash"
	"strings"
	"sync"
	"sync/atomic"
)

// UploadBatches coordinates active upload requests with batch expiry.
var UploadBatches sync.RWMutex

// UploadRecordSlots is the fixed number of request records that upload batch names are hashed to.
const UploadRecordSlots = 1024

// UploadRecords is a snapshot of the request records, taken before batch expiry scans for candidates.
type UploadRecords [UploadRecordSlots]uint64

// uploadRecords change whenever an upload or processing request begins or ends its work on a batch
// hashed to them, so batch expiry can tell whether a batch may have changed since its scan.
var uploadRecords [UploadRecordSlots]atomic.Uint64

// uploadRecordSeed is the process-local seed for hashing batch names to records.
var uploadRecordSeed = maphash.MakeSeed()

// unlockUploadBatches releases a request's shared lifecycle lock; tests replace it to observe the order.
var unlockUploadBatches = UploadBatches.RUnlock

// uploadRecordSlot returns the index of the record for the batch name, ignoring case so that names a
// case-insensitive file system resolves to the same folder share a record.
func uploadRecordSlot(batch string) uint64 {
	return maphash.String(uploadRecordSeed, strings.ToLower(batch)) % UploadRecordSlots
}

// UploadRecord returns the current request record of the batch name.
func UploadRecord(batch string) uint64 {
	return uploadRecords[uploadRecordSlot(batch)].Load()
}

// LoadUploadRecords returns a snapshot of all request records.
func LoadUploadRecords() (records UploadRecords) {
	for i := range uploadRecords {
		records[i] = uploadRecords[i].Load()
	}
	return records
}

// Get returns the record of the batch name at the time of the snapshot.
func (records *UploadRecords) Get(batch string) uint64 {
	return records[uploadRecordSlot(batch)]
}

// BeginUploadRequest holds the batch lifecycle lock shared and records the request for the batch name.
func BeginUploadRequest(batch string) {
	UploadBatches.RLock()
	uploadRecords[uploadRecordSlot(batch)].Add(1)
}

// EndUploadRequest records the end of a request for the batch name and releases its shared lock. The
// record comes first, so expiry that acquires the lock afterwards sees that the request ran after its scan.
func EndUploadRequest(batch string) {
	uploadRecords[uploadRecordSlot(batch)].Add(1)
	unlockUploadBatches()
}
