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

func TestError(t *testing.T) {
	s := Subscribe("notify.error")

	assert.IsType(t, hub.Subscription{}, s)

	Error("error message")
	msg := <-s.Receiver

	assert.Equal(t, "notify.error", msg.Name)
	assert.Equal(t, Data{"message": "error message"}, msg.Fields)

	Unsubscribe(s)
}

func TestSuccess(t *testing.T) {
	s := Subscribe("notify.success")

	assert.IsType(t, hub.Subscription{}, s)

	Success("success message")
	msg := <-s.Receiver

	assert.Equal(t, "notify.success", msg.Name)
	assert.Equal(t, Data{"message": "success message"}, msg.Fields)

	Unsubscribe(s)
}

func TestInfo(t *testing.T) {
	s := Subscribe("notify.info")

	assert.IsType(t, hub.Subscription{}, s)

	Info("info message")
	msg := <-s.Receiver

	assert.Equal(t, "notify.info", msg.Name)
	assert.Equal(t, Data{"message": "info message"}, msg.Fields)

	Unsubscribe(s)
}

func TestWarning(t *testing.T) {
	s := Subscribe("notify.warning")

	assert.IsType(t, hub.Subscription{}, s)

	Warning("warning message")
	msg := <-s.Receiver

	assert.Equal(t, "notify.warning", msg.Name)
	assert.Equal(t, Data{"message": "warning message"}, msg.Fields)

	Unsubscribe(s)
}
