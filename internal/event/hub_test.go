package event

import (
	"testing"

	"github.com/leandro-lugaresi/hub"
	"github.com/stretchr/testify/assert"
)

func TestSharedHub(t *testing.T) {
	h := SharedHub()

	assert.IsType(t, &hub.Hub{}, h)
}

func TestPublishSubscribe(t *testing.T) {
	s := Subscribe("foo.bar")

	assert.IsType(t, hub.Subscription{}, s)

	Publish("foo.bar", Data{"id": 13})

	msg := <-s.Receiver

    t.Logf("receive msg with topic %s: %v\n", msg.Name, msg.Fields)

	assert.Equal(t, "foo.bar", msg.Name)
	assert.Equal(t, Data{"id": 13}, msg.Fields)

	Unsubscribe(s)
}
