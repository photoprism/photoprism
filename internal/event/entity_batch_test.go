package event

import (
	"testing"
	"time"

	"github.com/leandro-lugaresi/hub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// subscribeEntityBatch subscribes to a test topic for the duration of a test.
func subscribeEntityBatch(t *testing.T, topic string) hub.Subscription {
	t.Helper()

	sub := Subscribe(topic)
	t.Cleanup(func() { Unsubscribe(sub) })

	return sub
}

// receiveEntityBatches drains the subscription and returns the identity list of each event received.
func receiveEntityBatches(t *testing.T, sub hub.Subscription) (events [][]string) {
	t.Helper()

	for {
		select {
		case msg := <-sub.Receiver:
			ids, ok := msg.Fields["entities"].([]string)
			require.True(t, ok, "entities payload should be []string, got %T", msg.Fields["entities"])
			events = append(events, ids)
		case <-time.After(250 * time.Millisecond):
			return events
		}
	}
}

func TestNewEntityBatch(t *testing.T) {
	b := NewEntityBatch("batch", EntityDeleted)

	require.NotNil(t, b)
	assert.Equal(t, "batch", b.channel)
	assert.Equal(t, EntityDeleted, b.action)
	assert.Equal(t, EntityBatchSize, b.maxSize)
	assert.Equal(t, EntityBatchWait, b.maxWait)
	assert.Empty(t, b.entities)
}

func TestEntityBatch_Add(t *testing.T) {
	const id1, id2, id3 = "pqkm36fjqvset9uy", "pt3cs9f5kvfxxvta", "ps6sg6be2lvl0yh7"

	t.Run("FirstAddPublishes", func(t *testing.T) {
		sub := subscribeEntityBatch(t, "batch.updated")
		b := NewEntityBatch("batch", EntityUpdated)

		b.Add(id1)

		assert.Equal(t, [][]string{{id1}}, receiveEntityBatches(t, sub))
		assert.Empty(t, b.entities)
	})
	t.Run("HoldsBatchWithinWait", func(t *testing.T) {
		sub := subscribeEntityBatch(t, "batch.updated")
		b := NewEntityBatch("batch", EntityUpdated)

		b.Add(id1)
		b.Add(id2)
		b.FlushDue()
		assert.Equal(t, [][]string{{id1}}, receiveEntityBatches(t, sub))

		b.published = b.published.Add(-b.maxWait)
		b.Add(id3)
		assert.Equal(t, [][]string{{id2, id3}}, receiveEntityBatches(t, sub))
	})
	t.Run("BatchesUntilFull", func(t *testing.T) {
		sub := subscribeEntityBatch(t, "batch.updated")
		b := &EntityBatch{channel: "batch", action: EntityUpdated, maxSize: 3, maxWait: time.Hour, published: time.Now()}

		b.Add(id1)
		b.Add(id2)
		assert.Empty(t, receiveEntityBatches(t, sub))

		b.Add(id3)
		assert.Equal(t, [][]string{{id1, id2, id3}}, receiveEntityBatches(t, sub))
	})
	t.Run("PublishesAfterWait", func(t *testing.T) {
		sub := subscribeEntityBatch(t, "batch.updated")
		b := &EntityBatch{channel: "batch", action: EntityUpdated, maxSize: 50, maxWait: time.Hour, published: time.Now().Add(-2 * time.Hour)}

		b.Add(id1)

		assert.Equal(t, [][]string{{id1}}, receiveEntityBatches(t, sub))
	})
	t.Run("Action", func(t *testing.T) {
		sub := subscribeEntityBatch(t, "batch.deleted")
		b := NewEntityBatch("batch", EntityDeleted)

		b.Add(id1)

		assert.Equal(t, [][]string{{id1}}, receiveEntityBatches(t, sub))
	})
	t.Run("SkipsEmptyAndDuplicate", func(t *testing.T) {
		sub := subscribeEntityBatch(t, "batch.updated")
		b := &EntityBatch{channel: "batch", action: EntityUpdated, maxSize: 50, maxWait: time.Hour, published: time.Now()}

		b.Add("")
		b.Add(id1)
		b.Add(id1)
		b.Flush()

		assert.Equal(t, [][]string{{id1}}, receiveEntityBatches(t, sub))
	})
	t.Run("Nil", func(t *testing.T) {
		var b *EntityBatch

		assert.NotPanics(t, func() {
			b.Add(id1)
			b.FlushDue()
			b.Flush()
		})
	})
}

func TestEntityBatch_FlushDue(t *testing.T) {
	const id = "pqkm36fjqvset9uy"

	t.Run("Due", func(t *testing.T) {
		sub := subscribeEntityBatch(t, "batch.updated")
		b := &EntityBatch{channel: "batch", action: EntityUpdated, entities: []string{id}, maxSize: 50, maxWait: time.Second, published: time.Now().Add(-time.Minute)}

		b.FlushDue()

		assert.Equal(t, [][]string{{id}}, receiveEntityBatches(t, sub))
		assert.Empty(t, b.entities)
	})
	t.Run("NotDue", func(t *testing.T) {
		sub := subscribeEntityBatch(t, "batch.updated")
		b := &EntityBatch{channel: "batch", action: EntityUpdated, entities: []string{id}, maxSize: 50, maxWait: time.Hour, published: time.Now()}

		b.FlushDue()

		assert.Empty(t, receiveEntityBatches(t, sub))
		assert.Equal(t, []string{id}, b.entities)
	})
	t.Run("Empty", func(t *testing.T) {
		sub := subscribeEntityBatch(t, "batch.updated")
		b := NewEntityBatch("batch", EntityUpdated)

		b.FlushDue()

		assert.Empty(t, receiveEntityBatches(t, sub))
	})
}

func TestEntityBatch_Flush(t *testing.T) {
	const id1, id2 = "pqkm36fjqvset9uy", "pt3cs9f5kvfxxvta"

	t.Run("PublishesPending", func(t *testing.T) {
		sub := subscribeEntityBatch(t, "batch.updated")
		b := &EntityBatch{channel: "batch", action: EntityUpdated, maxSize: 50, maxWait: time.Hour, published: time.Now()}

		b.Add(id1)
		b.Add(id2)
		b.Flush()
		b.Flush()

		assert.Equal(t, [][]string{{id1, id2}}, receiveEntityBatches(t, sub))
		assert.WithinDuration(t, time.Now(), b.published, time.Second)
	})
	t.Run("Empty", func(t *testing.T) {
		sub := subscribeEntityBatch(t, "batch.updated")
		b := NewEntityBatch("batch", EntityUpdated)

		b.Flush()

		assert.Empty(t, receiveEntityBatches(t, sub))
		assert.True(t, b.published.IsZero())
	})
}

func TestEntityBatch_flush(t *testing.T) {
	const id = "pqkm36fjqvset9uy"

	t.Run("Success", func(t *testing.T) {
		sub := subscribeEntityBatch(t, "batch.updated")
		b := &EntityBatch{channel: "batch", action: EntityUpdated, entities: []string{id}, maxSize: 50, maxWait: time.Hour}

		b.flush()

		assert.Equal(t, [][]string{{id}}, receiveEntityBatches(t, sub))
		assert.Nil(t, b.entities)
		assert.False(t, b.published.IsZero())
	})
	t.Run("NothingPending", func(t *testing.T) {
		sub := subscribeEntityBatch(t, "batch.updated")
		b := &EntityBatch{channel: "batch", action: EntityUpdated, maxSize: 50, maxWait: time.Hour}

		b.flush()

		assert.Empty(t, receiveEntityBatches(t, sub))
		assert.True(t, b.published.IsZero())
	})
}
