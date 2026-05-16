package photoprism

import (
	"sort"
	"sync"
	"time"
)

var defaultProgressWorkers = []string{ActionImport, ActionIndex}
var workerProgress = newWorkerProgressTracker()

// WorkerProgress contains observable file processing progress for a worker.
type WorkerProgress struct {
	Worker     string
	Running    bool
	StartedAt  time.Time
	FinishedAt time.Time
	Files      int
	Bytes      int64
}

type workerProgressTracker struct {
	mu        sync.RWMutex
	snapshots map[string]WorkerProgress
}

// newWorkerProgressTracker creates a tracker with stable default workers.
func newWorkerProgressTracker() *workerProgressTracker {
	result := &workerProgressTracker{
		snapshots: make(map[string]WorkerProgress, len(defaultProgressWorkers)),
	}

	for _, worker := range defaultProgressWorkers {
		result.snapshots[worker] = WorkerProgress{Worker: worker}
	}

	return result
}

// resetWorkerProgressForTest resets progress state for package tests.
func resetWorkerProgressForTest() {
	workerProgress = newWorkerProgressTracker()
}

// StartWorkerProgress marks a worker as running and resets the current run counters.
func StartWorkerProgress(worker string) {
	workerProgress.Start(worker)
}

// ObserveWorkerProgress records one processed file for the worker.
func ObserveWorkerProgress(worker string, bytes int64) {
	workerProgress.Observe(worker, bytes)
}

// FinishWorkerProgress marks a worker as idle while preserving its last run counters.
func FinishWorkerProgress(worker string) {
	workerProgress.Finish(worker)
}

// WorkerProgressSnapshot returns a copy of the current progress for one worker.
func WorkerProgressSnapshot(worker string) WorkerProgress {
	return workerProgress.Snapshot(worker)
}

// WorkerProgressSnapshots returns sorted copies of all worker progress snapshots.
func WorkerProgressSnapshots() []WorkerProgress {
	return workerProgress.Snapshots()
}

// Start marks a worker as running and resets the current run counters.
func (t *workerProgressTracker) Start(worker string) {
	if worker == "" {
		return
	}

	now := time.Now()

	t.mu.Lock()
	defer t.mu.Unlock()

	t.snapshots[worker] = WorkerProgress{
		Worker:    worker,
		Running:   true,
		StartedAt: now,
	}
}

// Observe records one processed file and its byte size for a worker.
func (t *workerProgressTracker) Observe(worker string, bytes int64) {
	if worker == "" {
		return
	}

	now := time.Now()

	t.mu.Lock()
	defer t.mu.Unlock()

	snapshot := t.snapshots[worker]

	if snapshot.Worker == "" {
		snapshot.Worker = worker
	}

	if snapshot.StartedAt.IsZero() {
		snapshot.StartedAt = now
	}

	snapshot.Running = true
	snapshot.FinishedAt = time.Time{}
	snapshot.Files++

	if bytes > 0 {
		snapshot.Bytes += bytes
	}

	t.snapshots[worker] = snapshot
}

// Finish marks a worker as idle without clearing its last run counters.
func (t *workerProgressTracker) Finish(worker string) {
	if worker == "" {
		return
	}

	now := time.Now()

	t.mu.Lock()
	defer t.mu.Unlock()

	snapshot := t.snapshots[worker]

	if snapshot.Worker == "" {
		snapshot.Worker = worker
	}

	snapshot.Running = false
	snapshot.FinishedAt = now
	t.snapshots[worker] = snapshot
}

// Snapshot returns a copy of the current progress for one worker.
func (t *workerProgressTracker) Snapshot(worker string) WorkerProgress {
	if worker == "" {
		return WorkerProgress{}
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	snapshot := t.snapshots[worker]

	if snapshot.Worker == "" {
		snapshot.Worker = worker
	}

	return snapshot
}

// Snapshots returns sorted copies of all tracked worker progress snapshots.
func (t *workerProgressTracker) Snapshots() []WorkerProgress {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]WorkerProgress, 0, len(t.snapshots))

	for worker, snapshot := range t.snapshots {
		if snapshot.Worker == "" {
			snapshot.Worker = worker
		}

		result = append(result, snapshot)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Worker < result[j].Worker
	})

	return result
}
