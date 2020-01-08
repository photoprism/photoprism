package mutex

import (
	"errors"
	"sync"
)

type Busy struct {
	busy     bool
	canceled bool
	mutex    sync.Mutex
}

func (b *Busy) Busy() bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return b.busy
}

func (b *Busy) Start() error {
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

func (b *Busy) Stop() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.busy = false
	b.canceled = false
}

func (b *Busy) Cancel() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.busy {
		b.canceled = true
	}
}

func (b *Busy) Canceled() bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return b.canceled
}
