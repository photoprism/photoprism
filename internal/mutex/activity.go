package mutex

import (
	"errors"
	"sync"
	"time"
)

// Activity represents work that can be started and stopped.
type Activity struct {
	busy     bool
	canceled bool
	mutex    sync.Mutex
	lastRun  time.Time
}

// Running checks if the Activity is currently running.
func (b *Activity) Running() bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return b.busy
}

// Start marks the Activity as started and returns an error if it is already running.
func (b *Activity) Start() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.canceled {
		return errors.New("still running")
	}

	if b.busy {
		return errors.New("already running")
	}

	b.busy = true
	b.canceled = false

	return nil
}

// Stop marks the Activity as stopped.
func (b *Activity) Stop() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.busy = false
	b.canceled = false
	b.lastRun = time.Now().UTC()
}

// Cancel requests to stop the Activity.
func (b *Activity) Cancel() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.busy {
		b.canceled = true
	}
}

// Canceled marks the Activity as stopped.
func (b *Activity) Canceled() bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return b.canceled
}

// LastRun returns the time of last activity.
func (b *Activity) LastRun() time.Time {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return b.lastRun
}
