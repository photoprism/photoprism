package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkerProgressSnapshotTracksCurrentRun(t *testing.T) {
	resetWorkerProgressForTest()

	StartWorkerProgress(ActionIndex)
	ObserveWorkerProgress(ActionIndex, 1024)
	ObserveWorkerProgress(ActionIndex, 2048)

	snapshot := WorkerProgressSnapshot(ActionIndex)

	assert.Equal(t, ActionIndex, snapshot.Worker)
	assert.True(t, snapshot.Running)
	assert.False(t, snapshot.StartedAt.IsZero())
	assert.True(t, snapshot.FinishedAt.IsZero())
	assert.Equal(t, 2, snapshot.Files)
	assert.Equal(t, int64(3072), snapshot.Bytes)

	FinishWorkerProgress(ActionIndex)

	snapshot = WorkerProgressSnapshot(ActionIndex)

	assert.False(t, snapshot.Running)
	assert.False(t, snapshot.StartedAt.IsZero())
	assert.False(t, snapshot.FinishedAt.IsZero())
	assert.Equal(t, 2, snapshot.Files)
	assert.Equal(t, int64(3072), snapshot.Bytes)
}

func TestWorkerProgressSnapshotsIncludeDefaultWorkers(t *testing.T) {
	resetWorkerProgressForTest()

	snapshots := WorkerProgressSnapshots()

	assert.Len(t, snapshots, 2)
	assert.Equal(t, ActionImport, snapshots[0].Worker)
	assert.Equal(t, ActionIndex, snapshots[1].Worker)
}
