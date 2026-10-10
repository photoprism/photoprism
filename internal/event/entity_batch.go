package event

import (
	"slices"
	"sync"
	"time"
)

const (
	// EntityBatchSize matches the number of photos the web UI refreshes in place per event.
	EntityBatchSize = 50
	// EntityBatchWait is the minimum time between two publishes of a batch that is not yet full.
	EntityBatchWait = 2 * time.Second
)

// EntityBatch collects the identities of changed entities and publishes them as <channel>.<action> events.
// Batching limits the event volume of fast background runs, while slow runs still notify about once per entity.
// Create it with NewEntityBatch, since the zero value publishes on every Add.
type EntityBatch struct {
	mu        sync.Mutex
	channel   string
	action    string
	entities  []string
	published time.Time
	maxSize   int
	maxWait   time.Duration
}

// NewEntityBatch returns an EntityBatch for the channel and action with the default size and wait limits.
func NewEntityBatch(channel, action string) *EntityBatch {
	return &EntityBatch{channel: channel, action: action, maxSize: EntityBatchSize, maxWait: EntityBatchWait}
}

// Add records a changed entity and publishes the pending identities once the batch is full or due.
func (b *EntityBatch) Add(id string) {
	if b == nil || id == "" {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if slices.Contains(b.entities, id) {
		return
	}

	b.entities = append(b.entities, id)

	if len(b.entities) >= b.maxSize || time.Since(b.published) >= b.maxWait {
		b.flush()
	}
}

// FlushDue publishes the pending identities if the wait has passed since the last publish.
// Callers in a loop invoke it per iteration, so a change followed by slow no-op iterations is not held back.
func (b *EntityBatch) FlushDue() {
	if b == nil {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if time.Since(b.published) >= b.maxWait {
		b.flush()
	}
}

// Flush publishes the pending identities as a single event.
func (b *EntityBatch) Flush() {
	if b == nil {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.flush()
}

// flush publishes the pending identities; the caller must hold the lock.
func (b *EntityBatch) flush() {
	if len(b.entities) == 0 {
		return
	}

	PublishEntities(b.channel, b.action, b.entities)

	b.entities = nil
	b.published = time.Now()
}
