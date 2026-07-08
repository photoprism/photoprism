package entity

import (
	"sync"
	"testing"
	"time"

	"github.com/photoprism/photoprism/pkg/time/unix"
)

func TestShouldUpdateLabelCounts(t *testing.T) {
	prev := updateLabelCountsLastUpdated.Load()
	defer updateLabelCountsLastUpdated.Store(prev)

	updateLabelCountsLastUpdated.Store(0)
	if !ShouldUpdateLabelCounts() {
		t.Fatalf("expected true when never run")
	}

	recent := unix.Now()
	updateLabelCountsLastUpdated.Store(recent)
	if ShouldUpdateLabelCounts() {
		t.Fatalf("expected false within default interval")
	}

	prevInterval := UpdateLabelCountsInterval
	UpdateLabelCountsInterval = 1
	defer func() { UpdateLabelCountsInterval = prevInterval }()

	updateLabelCountsLastUpdated.Store(unix.Now() - 5)
	if !ShouldUpdateLabelCounts() {
		t.Fatalf("expected true after interval elapsed")
	}
}

func TestUpdateLabelCountsIfNeeded(t *testing.T) {
	prev := updateLabelCountsLastUpdated.Load()
	defer updateLabelCountsLastUpdated.Store(prev)

	recent := unix.Now()
	updateLabelCountsLastUpdated.Store(recent)
	if err := UpdateLabelCountsIfNeeded(); err != nil {
		t.Fatalf("expected nil when skipping update, got %v", err)
	}
	if updateLabelCountsLastUpdated.Load() != recent {
		t.Fatalf("timestamp should remain unchanged when skipping update")
	}

	prevInterval := UpdateLabelCountsInterval
	UpdateLabelCountsInterval = 0
	defer func() { UpdateLabelCountsInterval = prevInterval }()

	updateLabelCountsLastUpdated.Store(0)
	before := time.Now()
	if err := UpdateLabelCountsIfNeeded(); err != nil {
		t.Fatalf("expected update to succeed, got %v", err)
	}
	after := updateLabelCountsLastUpdated.Load()
	if after == 0 {
		t.Fatalf("expected timestamp to be recorded")
	}
	if time.Unix(after, 0).Before(before.Add(-time.Minute)) {
		t.Fatalf("timestamp not refreshed")
	}
}

func TestUpdateCounts_NilDbReturnsCleanly(t *testing.T) {
	// Simulate the post-CloseDb shutdown state where the entity DB
	// provider has been nilled. UpdateCounts must return nil instead of
	// panicking on a nil dialect lookup, otherwise an in-flight async
	// goroutine spawned by UpdateCountsAsync would crash the process.
	prev := dbConn
	defer SetDbProvider(prev)
	SetDbProvider(nil)

	if err := UpdateCounts(); err != nil {
		t.Fatalf("UpdateCounts on nil DB should return nil, got %v", err)
	}
}

func TestWaitForAsyncJobs_DrainsRegisteredWork(t *testing.T) {
	// Models the contract that config.CloseDb relies on: WaitForAsyncJobs
	// must block until every AsyncJobAdd has a matching AsyncJobDone, so
	// async count/cover update goroutines finish before the DB connection
	// is torn down.
	const workers = 8

	started := make(chan struct{}, workers)
	release := make(chan struct{})
	var done sync.WaitGroup

	done.Add(workers)
	for range workers {
		AsyncJobAdd()
		go func() {
			defer AsyncJobDone()
			defer done.Done()
			started <- struct{}{}
			<-release
		}()
	}

	// Make sure every goroutine is parked before we start the wait so the
	// test exercises a real drain rather than a fast path on an empty WG.
	for range workers {
		<-started
	}

	waitDone := make(chan struct{})
	go func() {
		WaitForAsyncJobs()
		close(waitDone)
	}()

	select {
	case <-waitDone:
		t.Fatalf("WaitForAsyncJobs returned before any worker called AsyncJobDone")
	case <-time.After(20 * time.Millisecond):
		// Expected — workers are still parked.
	}

	close(release)
	done.Wait()

	select {
	case <-waitDone:
		// Expected.
	case <-time.After(2 * time.Second):
		t.Fatalf("WaitForAsyncJobs did not return after all workers finished")
	}
}
