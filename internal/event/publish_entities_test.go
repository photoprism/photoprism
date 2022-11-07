package event

import (
	"testing"

	"github.com/leandro-lugaresi/hub"
	"github.com/stretchr/testify/assert"
)

func TestEntitiesUpdated(t *testing.T) {
	s := Subscribe("test.updated")

	assert.IsType(t, hub.Subscription{}, s)

	entities := "test"
	EntitiesUpdated("test", entities)
	msg := <-s.Receiver

	assert.Equal(t, "test.updated", msg.Name)
	assert.Equal(t, Data{"entities": "test"}, msg.Fields)

	Unsubscribe(s)
}

func TestEntitiesCreated(t *testing.T) {
	s := Subscribe("test.created")

	assert.IsType(t, hub.Subscription{}, s)

	entities := "test"
	EntitiesCreated("test", entities)
	msg := <-s.Receiver

	assert.Equal(t, "test.created", msg.Name)
	assert.Equal(t, Data{"entities": "test"}, msg.Fields)

	Unsubscribe(s)
}

func TestEntitiesDeleted(t *testing.T) {
	s := Subscribe("test.deleted")

	assert.IsType(t, hub.Subscription{}, s)

	entities := "test"
	EntitiesDeleted("test", entities)
	msg := <-s.Receiver

	assert.Equal(t, "test.deleted", msg.Name)
	assert.Equal(t, Data{"entities": "test"}, msg.Fields)

	Unsubscribe(s)
}

func TestEntitiesArchived(t *testing.T) {
	s := Subscribe("test.archived")

	assert.IsType(t, hub.Subscription{}, s)

	entities := "test"
	EntitiesArchived("test", entities)
	msg := <-s.Receiver

	assert.Equal(t, "test.archived", msg.Name)
	assert.Equal(t, Data{"entities": "test"}, msg.Fields)

	Unsubscribe(s)
}

func TestEntitiesRestored(t *testing.T) {
	s := Subscribe("test.restored")

	assert.IsType(t, hub.Subscription{}, s)

	entities := "test"
	EntitiesRestored("test", entities)
	msg := <-s.Receiver

	assert.Equal(t, "test.restored", msg.Name)
	assert.Equal(t, Data{"entities": "test"}, msg.Fields)

	Unsubscribe(s)
}
