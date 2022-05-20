package event

import (
	"bytes"
	"sync"
)

// Buffer is a goroutine safe buffer.
type Buffer struct {
	buffer bytes.Buffer
	mutex  sync.RWMutex
}

// Set updates the buffer content.
func (b *Buffer) Set(s string) (err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.buffer.Reset()
	_, err = b.buffer.WriteString(s)
	return err
}

// Get returns the buffer content.
func (b *Buffer) Get() string {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.buffer.String()
}
