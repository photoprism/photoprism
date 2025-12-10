package event

import (
	"strings"

	"github.com/leandro-lugaresi/hub"
)

// Hub is an alias for the shared event hub implementation.
type Hub = hub.Hub

// Data represents event payload fields.
type Data = hub.Fields

// Message represents an emitted event message.
type Message = hub.Message

// TopicSep separates topic hierarchy segments.
const TopicSep = "."

var channelCap = 100
var sharedHub = NewHub()

// NewHub creates a new event hub instance.
func NewHub() *Hub {
	return hub.New()
}

// SharedHub returns the process-wide shared event hub.
func SharedHub() *Hub {
	return sharedHub
}

// Subscribe creates a topic subscription and returns i
func Subscribe(topics ...string) hub.Subscription {
	return SharedHub().NonBlockingSubscribe(channelCap, topics...)
}

// Unsubscribe deletes the subscription of a topic.
func Unsubscribe(s hub.Subscription) {
	SharedHub().Unsubscribe(s)
}

// Topic splits the topic name into the channel and event names.
func Topic(topic string) (ch, ev string) {
	ch, ev, _ = strings.Cut(topic, TopicSep)
	return ch, ev
}
