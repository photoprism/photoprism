package mutex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// uploadLockState reports whether expiry could take the lifecycle lock and whether a request could share it.
func uploadLockState() (exclusive, shared bool) {
	if exclusive = UploadBatches.TryLock(); exclusive {
		UploadBatches.Unlock()
	}
	if shared = UploadBatches.TryRLock(); shared {
		UploadBatches.RUnlock()
	}
	return exclusive, shared
}

// TestBeginUploadRequest verifies that a request holds the lifecycle lock shared and is recorded.
func TestBeginUploadRequest(t *testing.T) {
	before := UploadRequests.Load()
	BeginUploadRequest()
	t.Cleanup(EndUploadRequest)
	assert.Equal(t, before+1, UploadRequests.Load())
	exclusive, shared := uploadLockState()
	assert.False(t, exclusive, "a request must exclude expiry")
	assert.True(t, shared, "requests must admit other requests")
}

// TestEndUploadRequest verifies that the end of a request is recorded and releases the lock.
func TestEndUploadRequest(t *testing.T) {
	BeginUploadRequest()
	before := UploadRequests.Load()
	EndUploadRequest()
	assert.Equal(t, before+1, UploadRequests.Load())
	exclusive, _ := uploadLockState()
	assert.True(t, exclusive, "the lock must be released")
}

// TestEndUploadRequest_RecordBeforeUnlock verifies that the end of a request is recorded before the lock is released.
func TestEndUploadRequest_RecordBeforeUnlock(t *testing.T) {
	unlock := unlockUploadBatches
	t.Cleanup(func() { unlockUploadBatches = unlock })
	var atUnlock uint64
	unlockUploadBatches = func() {
		atUnlock = UploadRequests.Load()
		unlock()
	}
	BeginUploadRequest()
	before := UploadRequests.Load()
	EndUploadRequest()
	assert.Equal(t, before+1, atUnlock, "expiry acquiring the lock must see the recorded end")
	exclusive, _ := uploadLockState()
	assert.True(t, exclusive, "the lock must be released")
}
