package mutex

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// otherUploadBatch returns a batch name whose record differs from the record of the given name.
func otherUploadBatch(t *testing.T, batch string) string {
	t.Helper()
	for i := range 100 {
		if other := fmt.Sprintf("other%d", i); uploadRecordSlot(other) != uploadRecordSlot(batch) {
			return other
		}
	}
	require.FailNow(t, "no batch name with a different record")
	return ""
}

// TestBeginUploadRequest verifies that a request holds the lifecycle lock shared and is recorded.
func TestBeginUploadRequest(t *testing.T) {
	before := UploadRecord("batch")
	BeginUploadRequest("batch")
	t.Cleanup(func() { EndUploadRequest("batch") })
	assert.Equal(t, before+1, UploadRecord("batch"))
	exclusive, shared := uploadLockState()
	assert.False(t, exclusive, "a request must exclude expiry")
	assert.True(t, shared, "requests must admit other requests")
}

// TestEndUploadRequest verifies that the end of a request is recorded and releases the lock.
func TestEndUploadRequest(t *testing.T) {
	BeginUploadRequest("batch")
	before := UploadRecord("batch")
	EndUploadRequest("batch")
	assert.Equal(t, before+1, UploadRecord("batch"))
	exclusive, _ := uploadLockState()
	assert.True(t, exclusive, "the lock must be released")
}

// TestEndUploadRequest_RecordBeforeUnlock verifies that the end of a request is recorded before the lock is released.
func TestEndUploadRequest_RecordBeforeUnlock(t *testing.T) {
	unlock := unlockUploadBatches
	t.Cleanup(func() { unlockUploadBatches = unlock })
	var atUnlock uint64
	unlockUploadBatches = func() {
		atUnlock = UploadRecord("batch")
		unlock()
	}
	BeginUploadRequest("batch")
	before := UploadRecord("batch")
	EndUploadRequest("batch")
	assert.Equal(t, before+1, atUnlock, "expiry acquiring the lock must see the recorded end")
	exclusive, _ := uploadLockState()
	assert.True(t, exclusive, "the lock must be released")
}

// TestUploadRecord verifies that a request changes the record of its own batch only.
func TestUploadRecord(t *testing.T) {
	t.Run("SameBatch", func(t *testing.T) {
		before := UploadRecord("batch")
		BeginUploadRequest("batch")
		EndUploadRequest("batch")
		assert.Equal(t, before+2, UploadRecord("batch"))
	})
	t.Run("OtherBatch", func(t *testing.T) {
		other := otherUploadBatch(t, "batch")
		before := UploadRecord("batch")
		BeginUploadRequest(other)
		EndUploadRequest(other)
		assert.Equal(t, before, UploadRecord("batch"))
	})
	t.Run("CaseVariant", func(t *testing.T) {
		before := UploadRecord("Batch")
		BeginUploadRequest("BATCH")
		EndUploadRequest("BATCH")
		assert.Equal(t, before+2, UploadRecord("batch"))
	})
}

// TestUploadRecordSlot verifies that batch names map to a valid record that ignores case.
func TestUploadRecordSlot(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		slots := make(map[uint64]bool)
		for i := range 100 {
			slot := uploadRecordSlot(fmt.Sprintf("batch%d", i))
			assert.Less(t, slot, uint64(UploadRecordSlots))
			slots[slot] = true
		}
		assert.Greater(t, len(slots), 1, "names must be spread across records")
	})
	t.Run("CaseVariant", func(t *testing.T) {
		assert.Equal(t, uploadRecordSlot("abc123XYZ"), uploadRecordSlot("ABC123xyz"))
	})
}

// TestUploadRecords_Get verifies that a snapshot returns the record of a batch name.
func TestUploadRecords_Get(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var records UploadRecords
		records[uploadRecordSlot("batch")] = 42
		assert.Equal(t, uint64(42), records.Get("batch"))
		assert.Equal(t, uint64(42), records.Get("BATCH"))
	})
	t.Run("OtherBatch", func(t *testing.T) {
		var records UploadRecords
		records[uploadRecordSlot("batch")] = 42
		assert.Zero(t, records.Get(otherUploadBatch(t, "batch")))
	})
}

// TestLoadUploadRecords verifies that a snapshot keeps the records from the time it was taken.
func TestLoadUploadRecords(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		other := otherUploadBatch(t, "batch")
		records := LoadUploadRecords()
		assert.Equal(t, UploadRecord("batch"), records.Get("batch"))
		BeginUploadRequest("batch")
		EndUploadRequest("batch")
		assert.Equal(t, records.Get("batch")+2, UploadRecord("batch"))
		assert.Equal(t, records.Get(other), UploadRecord(other))
	})
	t.Run("AllSlots", func(t *testing.T) {
		BeginUploadRequest("batch")
		EndUploadRequest("batch")
		records := LoadUploadRecords()
		assert.NotZero(t, records.Get("batch"))
		for i := range uploadRecords {
			assert.Equal(t, uploadRecords[i].Load(), records[i])
		}
	})
}
