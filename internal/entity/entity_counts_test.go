package entity

import (
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
